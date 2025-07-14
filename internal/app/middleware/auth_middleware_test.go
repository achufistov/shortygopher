package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/achufistov/shortygopher.git/tests/testutils"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware_NewUser(t *testing.T) {
	cfg := testutils.CreateTestConfig(t, "test-secret-key-for-auth-middleware")

	// Create test handler that checks if userID is set in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			t.Error("Expected userID to be set in context")
		}

		userIDStr, ok := userID.(string)
		if !ok {
			t.Error("Expected userID to be string")
		}
		if userIDStr == "" {
			t.Error("Expected non-empty userID")
		}

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	middleware := AuthMiddleware(cfg)
	handler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check if auth cookie was set
	result := w.Result()
	if result != nil {
		defer result.Body.Close()
		cookies := result.Cookies()
		var authCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				authCookie = cookie
				break
			}
		}

		if authCookie == nil {
			t.Error("Expected auth_token cookie to be set")
		} else {
			if authCookie.Value == "" {
				t.Error("Expected non-empty auth_token value")
			}

			if !authCookie.HttpOnly {
				t.Error("Expected auth_token cookie to be HttpOnly")
			}

			if authCookie.MaxAge != 86400 {
				t.Errorf("Expected MaxAge 86400, got %d", authCookie.MaxAge)
			}
		}
	}
}

func TestAuthMiddleware_ExistingValidToken(t *testing.T) {
	cfg := testutils.CreateTestConfig(t, "test-secret-key-for-auth-middleware")

	// Create a valid JWT token
	testUserID := "test-user-123"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// Create test handler that checks userID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			t.Error("Expected userID to be set in context")
		}

		userIDStr, ok := userID.(string)
		if !ok {
			t.Error("Expected userID to be string")
		}
		if userIDStr != testUserID {
			t.Errorf("Expected userID '%s', got '%s'", testUserID, userIDStr)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	middleware := AuthMiddleware(cfg)
	handler := middleware(testHandler)

	// Create test request with auth cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that no new cookie was set (should use existing valid token)
	result := w.Result()
	if result != nil {
		defer result.Body.Close()
		cookies := result.Cookies()
		if len(cookies) > 0 {
			t.Error("Expected no new cookies to be set for existing valid token")
		}
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	cfg := testutils.CreateTestConfig(t, "test-secret-key-for-auth-middleware")

	// Create an expired JWT token
	testUserID := "test-user-123"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	})

	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// Create test handler that checks for new userID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			t.Error("Expected userID to be set in context")
		}

		userIDStr, ok := userID.(string)
		if !ok {
			t.Error("Expected userID to be string")
		}
		if userIDStr == "" {
			t.Error("Expected non-empty userID")
		}
		// Should be a new userID, not the one from expired token
		if userIDStr == testUserID {
			t.Error("Expected new userID for expired token")
		}

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	middleware := AuthMiddleware(cfg)
	handler := middleware(testHandler)

	// Create test request with expired auth cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check if new auth cookie was set
	result := w.Result()
	if result != nil {
		defer result.Body.Close()
		cookies := result.Cookies()
		var authCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				authCookie = cookie
				break
			}
		}

		if authCookie == nil {
			t.Error("Expected new auth_token cookie to be set for expired token")
		}
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	cfg := testutils.CreateTestConfig(t, "test-secret-key-for-auth-middleware")

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			t.Error("Expected userID to be set in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	middleware := AuthMiddleware(cfg)
	handler := middleware(testHandler)

	// Create test request with invalid auth cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "invalid.jwt.token",
	})
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check if new auth cookie was set
	result := w.Result()
	if result != nil {
		defer result.Body.Close()
		cookies := result.Cookies()
		var authCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				authCookie = cookie
				break
			}
		}

		if authCookie == nil {
			t.Error("Expected new auth_token cookie to be set for invalid token")
		}
	}
}

func TestAuthMiddleware_ContextKey(t *testing.T) {
	// Test that UserIDKey is properly defined
	if UserIDKey != "userID" {
		t.Errorf("Expected UserIDKey to be 'userID', got '%s'", UserIDKey)
	}

	// Test that context key can be used properly
	ctx := context.Background()
	testUserID := "test-user-123"

	ctx = context.WithValue(ctx, UserIDKey, testUserID)

	retrievedUserID := ctx.Value(UserIDKey)
	if retrievedUserID != testUserID {
		t.Errorf("Expected userID '%s', got '%v'", testUserID, retrievedUserID)
	}
}
