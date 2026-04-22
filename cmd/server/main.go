package main

import (
	"log"
	"net/http"
	"os"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, using system environment variables")
	}

	// Initialize database
	database.InitDB()
	defer database.DB.Close()

	if err := api.SetJWTSecret(os.Getenv("JWT_SECRET")); err != nil {
		log.Fatal("failed to configure JWT secret: ", err)
	}

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
