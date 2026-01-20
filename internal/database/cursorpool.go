package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"GoBI/internal/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type CursorState struct {
	CursorName string
	Conn       *sql.Conn
	Tx         *sql.Tx
	LastUsed   time.Time
	sync.Mutex
	PageSize int
}

type CursorPool struct {
	db          *sql.DB
	cursors     map[string]*CursorState
	mu          sync.Mutex
	maxConns    int
	idleTimeout time.Duration
}

func NewCursorPool(connStr string, cfg config.CursorPoolConfig) (*CursorPool, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	idleTimeout, _ := time.ParseDuration(cfg.IdleTimeout)
	if idleTimeout == 0 {
		idleTimeout = 30 * time.Second
	}

	pool := &CursorPool{
		db:          db,
		cursors:     make(map[string]*CursorState),
		maxConns:    cfg.MaxConnections,
		idleTimeout: idleTimeout,
	}

	go pool.cleanupRoutine()
	return pool, nil
}

func (p *CursorPool) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for id, state := range p.cursors {
			if now.Sub(state.LastUsed) > p.idleTimeout {
				log.Printf("Closing idle cursor: %s", id)
				state.Tx.Rollback()
				state.Conn.Close()
				delete(p.cursors, id)
			}
		}
		p.mu.Unlock()
	}
}

func (p *CursorPool) ExecuteQuery(ctx context.Context, sessionID, query string, pageSize int, params map[string]interface{}) ([]map[string]interface{}, error) {
	processedSQL := ProcessSQL(query, params)

	p.mu.Lock()
	state, exists := p.cursors[sessionID]
	if exists {
		state.Tx.Rollback()
		state.Conn.Close()
	}

	conn, err := p.db.Conn(ctx)
	if err != nil {
		p.mu.Unlock()
		return nil, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		conn.Close()
		p.mu.Unlock()
		return nil, err
	}

	cursorName := "cur_" + uuid.New().String()[:8]
	declareQuery := fmt.Sprintf("DECLARE %s SCROLL CURSOR FOR %s", cursorName, processedSQL)

	if _, err := tx.ExecContext(ctx, declareQuery); err != nil {
		tx.Rollback()
		conn.Close()
		p.mu.Unlock()
		return nil, fmt.Errorf("failed to declare cursor: %w", err)
	}

	state = &CursorState{
		CursorName: cursorName,
		Conn:       conn,
		Tx:         tx,
		LastUsed:   time.Now(),
		PageSize:   pageSize,
	}
	p.cursors[sessionID] = state
	p.mu.Unlock()

	return p.FetchPage(ctx, sessionID, "NEXT")
}

func (p *CursorPool) FetchPage(ctx context.Context, sessionID, direction string) ([]map[string]interface{}, error) {
	p.mu.Lock()
	state, ok := p.cursors[sessionID]
	p.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("no active session")
	}

	state.Lock()
	defer state.Unlock()
	state.LastUsed = time.Now()

	fetchSQL := ""
	switch direction {
	case "NEXT":
		fetchSQL = fmt.Sprintf("FETCH FORWARD %d FROM %s", state.PageSize, state.CursorName)
	case "PREV":
		fetchSQL = fmt.Sprintf("MOVE RELATIVE -%d FROM %s; FETCH FORWARD %d FROM %s", 2*state.PageSize, state.CursorName, state.PageSize, state.CursorName)
	case "FIRST":
		fetchSQL = fmt.Sprintf("MOVE ABSOLUTE 0 FROM %s; FETCH FORWARD %d FROM %s", state.CursorName, state.PageSize, state.CursorName)
	}

	rows, err := state.Tx.QueryContext(ctx, fetchSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		args := make([]interface{}, len(cols))
		for i := range values {
			args[i] = &values[i]
		}
		rows.Scan(args...)

		row := make(map[string]interface{})
		for i, name := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[name] = string(b)
			} else {
				row[name] = val
			}
		}
		results = append(results, row)
	}
	return results, nil
}
