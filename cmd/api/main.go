package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	// Internal packages
	"github.com/peekdylan/reflekt-api/internal/api"
	"github.com/peekdylan/reflekt-api/internal/database"

	// Third-party packages
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver — imported for side effects only
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Open and verify the database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to the database")

	// Initialize the API config with all shared dependencies
	cfg := &api.APIConfig{
		DB:           database.New(db),
		JwtSecret:    jwtSecret,
		AnthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	// Set up the HTTP router
	mux := http.NewServeMux()

	// Health check — verifies the API is running
	mux.HandleFunc("GET /v1/health", cfg.HandlerHealth)

	// Auth routes — public, no token required
	mux.HandleFunc("POST /v1/register", cfg.HandlerRegister)
	mux.HandleFunc("POST /v1/login", cfg.HandlerLogin)

	// Journal entry routes — all protected by auth middleware
	mux.HandleFunc("POST /v1/entries", cfg.MiddlewareAuth(cfg.HandlerCreateEntry))
	mux.HandleFunc("GET /v1/entries", cfg.MiddlewareAuth(cfg.HandlerGetEntries))
	mux.HandleFunc("DELETE /v1/entries/{id}", cfg.MiddlewareAuth(cfg.HandlerDeleteEntry))

	// Wrap the entire router with CORS middleware so the frontend can
	// communicate with the API during local development
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: api.MiddlewareCORS(mux),
	}

	fmt.Printf("Reflekt API listening on port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
