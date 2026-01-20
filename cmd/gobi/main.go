package main

import (
	"GoBI/internal/config"
	"GoBI/internal/database"
	"GoBI/internal/handlers"
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := database.NewCursorPool(cfg.Database.GetConnectStr(), cfg.CursorPool)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}

	// Database Health Check
	log.Printf("Performing database health check...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Printf("Database health check successful.")

	// Load Repository Metadata
	repo, err := config.LoadRepository("ui/repository.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load repository: %v", err)
	} else {
		handlers.SetRepository(repo)
	}

	handlers.SetPool(pool)
	handlers.SetDatabaseName(cfg.Database.Database)

	http.HandleFunc("/", handlers.DashboardHandler)
	http.HandleFunc("/reports", handlers.ReportsHandler)
	http.HandleFunc("/report", handlers.ReportDetailHandler)
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))

	log.Printf("GoBI Server starting on :%s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		log.Fatal(err)
	}
}
