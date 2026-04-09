package api

import (
	"encoding/json"
	"net/http"
	"time"

	// Internal packages
	"github.com/peekdylan/reflekt-api/internal/database"

	// Third-party packages
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// registerRequest defines the expected JSON body for the register endpoint.
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// loginRequest defines the expected JSON body for the login endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authResponse is what we return to the client after successful register/login.
// Notice we never return the hashed password — only safe fields.
type authResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

// HandlerRegister creates a new user account.
// It hashes the password with bcrypt before storing it in the database.
func (cfg *APIConfig) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into our registerRequest struct
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation — make sure required fields are present
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	// Hash the password using bcrypt — we never store plain text passwords
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Insert the new user into the database using our SQLC-generated query
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		Name:           req.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate a JWT token for the newly registered user
	token, err := generateJWT(user.ID.String(), cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusCreated, authResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Token:     token,
		CreatedAt: user.CreatedAt.String(),
	})
}

// HandlerLogin authenticates an existing user and returns a JWT token.
func (cfg *APIConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	// Decode the request body
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Look up the user by email
	user, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Return a generic message so we don't reveal whether the email exists
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare the provided password against the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate a JWT token for the authenticated user
	token, err := generateJWT(user.ID.String(), cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, authResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Token:     token,
		CreatedAt: user.CreatedAt.String(),
	})
}

// generateJWT creates a signed JWT token for a given user ID.
// Tokens expire after 24 hours — after that the user must log in again.
func generateJWT(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,                                // Subject — the user this token belongs to
		"exp": time.Now().Add(24 * time.Hour).Unix(), // Expiry — 24 hours from now
		"iat": time.Now().Unix(),                     // Issued at
	})

	return token.SignedString([]byte(secret))
}
