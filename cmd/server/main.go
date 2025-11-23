package main

import (
	"fmt"
	"log"
	"net/http"
	"pr-review-service/internal/config"
	"pr-review-service/internal/database"
	"pr-review-service/internal/handlers"
	"pr-review-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	svc := service.NewService(db)

	h := handlers.NewHandlers(svc)

	mux := http.NewServeMux()
	h.SetupRoutes(mux)

	fmt.Printf("Starting server on port %s\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
