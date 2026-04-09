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
