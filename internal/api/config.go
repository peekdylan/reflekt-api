package api

import "github.com/peekdylan/reflekt-api/internal/database"

// APIConfig holds all shared dependencies for the API.
// All HTTP handlers are methods on this struct so they can
// access the database, secrets, and any other shared state.
type APIConfig struct {
	DB           *database.Queries // SQLC-generated database query layer
	JwtSecret    string            // Secret used to sign and verify JWT tokens
	AnthropicKey string            // API key for Claude AI analysis
}
