package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	// Internal packages
	"github.com/peekdylan/reflekt-api/internal/database"

	// Third-party packages
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver — the underscore means we import it for its side effects only
)

// apiConfig holds all shared state and dependencies for the API.
// Handlers are methods on this struct so they can access the DB, secrets, etc.
type apiConfig struct {
	db        *database.Queries // SQLC-generated database query layer
	jwtSecret string            // Secret used to sign and verify JWT tokens
}

func main() {
	// Load environment variables from .env file
	// This makes local development easy without hardcoding secrets
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Read required environment variables
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

	// Open a connection to PostgreSQL using the lib/pq driver
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Ping the database to verify the connection is actually working
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to the database")

	// Create the apiConfig with all shared dependencies
	cfg := &apiConfig{
		db:        database.New(db), // Wrap the DB connection with SQLC-generated queries
		jwtSecret: jwtSecret,
	}

	// Set up the HTTP router
	mux := http.NewServeMux()

	// Health check endpoint — useful for monitoring and Docker health checks
	mux.HandleFunc("GET /v1/health", cfg.handlerHealth)

	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Reflekt API listening on port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

// handlerHealth responds with a simple JSON status message.
// Used to verify the API is running and reachable.
func (cfg *apiConfig) handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
