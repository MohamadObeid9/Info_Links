package main

import (
	"log"
	"net/http"
	"os"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/database"
)

func main() {
	// Initialize database
	database.InitDB()
	defer database.DB.Close()

	// Setup router
	handler := api.NewRouter()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Backend server is starting on port %s...", port)
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
