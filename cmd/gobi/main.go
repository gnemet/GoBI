package main

import (
	"GoBI/internal/config"
	"GoBI/internal/database"
	"GoBI/internal/handlers"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := database.NewCursorPool(cfg.Database.GetConnectStr(), cfg.CursorPool)
	if err != nil {
		log.Printf("Warning: Database connection failed (simulating): %v", err)
	}
	handlers.SetPool(pool)

	http.HandleFunc("/", handlers.DashboardHandler)
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))

	log.Printf("GoBI Server starting on :%s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		log.Fatal(err)
	}
}
