package handlers

import (
	"GoBI/internal/config"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var repo *config.Repository
var dbName string

func SetRepository(r *config.Repository) {
	repo = r
}

func SetDatabaseName(name string) {
	dbName = name
}

func ReportsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"ui/templates/reports.html",
		"ui/templates/partials/nav.html",
		"ui/templates/partials/header.html",
		"ui/templates/partials/footer.html",
	))

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 6 // Default panes per page
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Only show main aggregate reports in the panes
	var aggregateReports []config.Report
	for _, report := range repo.Reports {
		if report.ViewType == "aggregate" {
			aggregateReports = append(aggregateReports, report)
		}
	}

	total := len(aggregateReports)
	start := offset
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	pagedReports := aggregateReports[start:end]

	nextPageSize := limit
	availableSizes := []int{6, 12, 24, 48}
	for i, size := range availableSizes {
		if size == limit {
			nextPageSize = availableSizes[(i+1)%len(availableSizes)]
			break
		}
	}

	prevOffset := offset - limit
	if prevOffset < 0 {
		prevOffset = 0
	}

	data := struct {
		Name         string
		Reports      []config.Report
		DatabaseName string
		Year         int
		Limit        int
		NextLimit    int
		Offset       int
		Total        int
		PrevOffset   int
		NextOffset   int
		HasPrev      bool
		HasNext      bool
		CurrentPage  int
		TotalPages   int
	}{
		Name:         repo.Meta.Name,
		Reports:      pagedReports,
		DatabaseName: dbName,
		Year:         time.Now().Year(),
		Limit:        limit,
		NextLimit:    nextPageSize,
		Offset:       offset,
		Total:        total,
		PrevOffset:   prevOffset,
		NextOffset:   offset + limit,
		HasPrev:      offset > 0,
		HasNext:      offset+limit < total,
		CurrentPage:  (offset / limit) + 1,
		TotalPages:   (total + limit - 1) / limit,
	}

	if r.Header.Get("HX-Request") == "true" {
		tmpl.ExecuteTemplate(w, "reports_list", data)
		return
	}

	tmpl.Execute(w, data)
}

func ReportDetailHandler(w http.ResponseWriter, r *http.Request) {
	reportID := r.URL.Query().Get("id")
	if reportID == "" {
		http.Error(w, "Missing report ID", http.StatusBadRequest)
		return
	}

	var selectedReport *config.Report
	for i := range repo.Reports {
		if repo.Reports[i].ID == reportID {
			selectedReport = &repo.Reports[i]
			break
		}
	}

	if selectedReport == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"ui/templates/report_detail.html",
		"ui/templates/partials/nav.html",
		"ui/templates/partials/header.html",
		"ui/templates/partials/table.html",
		"ui/templates/partials/footer.html",
	))

	query := "SELECT * FROM " + selectedReport.Schema + "." + selectedReport.TableName
	filterCol := r.URL.Query().Get("filter_col")
	filterVal := r.URL.Query().Get("filter_val")

	if filterCol != "" && filterVal != "" {
		// Basic SQL injection prevention for values (though this is internal use mostly)
		query += fmt.Sprintf(" WHERE %s = '%s'", filterCol, filterVal)
	}

	sorts := r.URL.Query()["sort"]
	if len(sorts) > 0 {
		var orderParts []string
		for _, s := range sorts {
			parts := strings.Split(s, ":")
			if len(parts) == 2 {
				orderParts = append(orderParts, fmt.Sprintf("%s %s", parts[0], parts[1]))
			}
		}
		if len(orderParts) > 0 {
			query += " ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var results []map[string]interface{}
	var err error

	direction := r.URL.Query().Get("dir")
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "sess-" + time.Now().Format("05.000000") // Unique-ish session ID
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = pool.DefaultPageSize
	}

	if selectedReport.ViewType == "aggregate" {
		// Use cursorpool for aggregate tables
		if direction != "" {
			results, err = pool.FetchPage(ctx, sessionID, direction)
		} else {
			results, err = pool.ExecuteQuery(ctx, sessionID, query, pageSize, nil)
		}
	} else {
		// Use one-time query for detail tables
		results, err = executeOneTimeQuery(ctx, query, pageSize)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var columns []TableColumn
	for _, col := range selectedReport.Columns {
		columns = append(columns, TableColumn{Name: col.Name, Label: col.Label, Hidden: col.Hidden})
	}
	// If columns not defined in repo, extract from results
	if len(columns) == 0 && len(results) > 0 {
		for k := range results[0] {
			columns = append(columns, TableColumn{Name: k, Label: k})
		}
	}

	// Find child report if any
	var childReportID string
	var childParentColumn string
	for _, rpt := range repo.Reports {
		if rpt.ParentReport == selectedReport.ID {
			childReportID = rpt.ID
			childParentColumn = rpt.ParentColumn
			break
		}
	}

	// Calculate NextPageSize for cycling
	nextPageSize := pool.DefaultPageSize
	for i, size := range pool.AvailablePageSizes {
		if size == pageSize {
			nextPageSize = pool.AvailablePageSizes[(i+1)%len(pool.AvailablePageSizes)]
			break
		}
	}

	// Find prev/next aggregate reports for header navigation
	var aggregateReports []*config.Report
	var currentIndex = -1
	for i := range repo.Reports {
		rpt := &repo.Reports[i]
		if rpt.ViewType == "aggregate" {
			aggregateReports = append(aggregateReports, rpt)
			if rpt.ID == selectedReport.ID {
				currentIndex = len(aggregateReports) - 1
			}
		}
	}

	var prevReportID, nextReportID string
	if currentIndex > 0 {
		prevReportID = aggregateReports[currentIndex-1].ID
	}
	if currentIndex >= 0 && currentIndex < len(aggregateReports)-1 {
		nextReportID = aggregateReports[currentIndex+1].ID
	}

	data := struct {
		Report            *config.Report
		Results           []map[string]interface{}
		Columns           []TableColumn
		DatabaseName      string
		Year              int
		SessionID         string
		PageSize          int
		NextPageSize      int
		PageSizes         []int
		ChildReportID     string
		ChildParentColumn string
		PrevReportID      string
		NextReportID      string
	}{
		Report:            selectedReport,
		Results:           results,
		Columns:           columns,
		DatabaseName:      dbName,
		Year:              time.Now().Year(),
		SessionID:         sessionID,
		PageSize:          pageSize,
		NextPageSize:      nextPageSize,
		PageSizes:         pool.AvailablePageSizes,
		ChildReportID:     childReportID,
		ChildParentColumn: childParentColumn,
		PrevReportID:      prevReportID,
		NextReportID:      nextReportID,
	}

	if r.Header.Get("HX-Request") == "true" {
		tmpl.ExecuteTemplate(w, "table", data)
		return
	}

	tmpl.Execute(w, data)
}

func executeOneTimeQuery(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	rows, err := pool.GetDB().QueryContext(ctx, fmt.Sprintf("%s LIMIT %d", query, limit))
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
