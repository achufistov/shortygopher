package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

// UserIDKey is the context key used to store and retrieve user ID from request context.
// Used by authentication middleware to pass user information between handlers.
const UserIDKey contextKey = "userID"

// AuthMiddleware returns HTTP middleware that handles JWT-based authentication.
// Validates existing JWT tokens from cookies or creates new ones for unauthenticated users.
// Sets user ID in request context for downstream handlers to access.
//
// The middleware:
//   - Checks for existing auth_token cookie
//   - Validates JWT token if present
//   - Generates new JWT token and sets cookie for new users
//   - Adds user ID to request context using UserIDKey
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			cookie, err := r.Cookie("auth_token")

			if err == nil {
				token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(cfg.SecretKey), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if uid, ok := claims["user_id"].(string); ok {
							userID = uid
						}
					}
				}
			}

			if userID == "" {
				userID = uuid.NewString()
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": userID,
					"exp":     time.Now().Add(24 * time.Hour).Unix(),
				})

				tokenString, err := token.SignedString([]byte(cfg.SecretKey))
				if err != nil {
					http.Error(w, "Failed to generate token", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    tokenString,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   86400,
				})
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
