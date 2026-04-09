package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type for context keys to avoid collisions
// with other packages that might use the same context keys.
type contextKey string

const (
	// contextKeyUserID is the key used to store the authenticated user's ID in the request context
	contextKeyUserID contextKey = "userID"
)

// MiddlewareCORS adds the necessary headers to allow the React Native web app
// running on localhost:8081 to make requests to our API on localhost:8080.
// In production this would be locked down to your actual domain.
func MiddlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from the Expo web dev server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests — browsers send these before the real request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareAuth protects routes that require a logged-in user.
// It reads the JWT token from the Authorization header, validates it,
// and injects the user ID into the request context for handlers to use.
func (cfg *APIConfig) MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The Authorization header should look like: "Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Split the header into "Bearer" and the token string
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate the JWT token using our secret
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Make sure the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JwtSecret), nil
		})
		if err != nil || !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Extract the user ID from the token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Invalid token subject")
			return
		}

		// Store the user ID in the request context so handlers can access it
		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next(w, r.WithContext(ctx))
	}
}

// getUserIDFromContext retrieves the authenticated user's ID from the request context.
// This is a helper used by protected handlers to know who is making the request.
func getUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	return userID, ok
}
