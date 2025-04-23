package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/google/uuid"
)

const (
	cookieName = "user_id"
)

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil || !ValidateCookie(cookie, cfg.SecretKey) {
				userID := uuid.New().String()
				cookie = createCookie(userID, cfg.SecretKey)
				http.SetCookie(w, cookie)
				r = r.WithContext(context.WithValue(r.Context(), "userID", userID))
			} else {
				userID := strings.Split(cookie.Value, ".")[0]
				r = r.WithContext(context.WithValue(r.Context(), "userID", userID))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func createCookie(userID, secretKey string) *http.Cookie {
	signature := signUserID(userID, secretKey)
	value := userID + "." + signature
	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
}

func ValidateCookie(cookie *http.Cookie, secretKey string) bool {
	if cookie == nil {
		return false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return false
	}
	userID, signature := parts[0], parts[1]
	return signature == signUserID(userID, secretKey)
}

func signUserID(userID, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
