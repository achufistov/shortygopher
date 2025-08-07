package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/service"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func TestStatsEndpointIntegration(t *testing.T) {
	// Create test configuration with only the field we need
	cfg := &config.Config{
		TrustedSubnet: "192.168.1.0/24",
	}

	// Create test storage
	testStorage := storage.NewURLStorage()

	// Add some test data
	testStorage.AddURL("abc123", "https://example1.com", "user1")
	testStorage.AddURL("def456", "https://example2.com", "user1")
	testStorage.AddURL("ghi789", "https://example3.com", "user2")
	testStorage.AddURL("jkl012", "https://example4.com", "user3")

	// Initialize handlers
	testService := service.NewService(testStorage, cfg)
	handlers.InitStorage(testStorage)
	handlers.InitService(testService)

	// Create router
	r := chi.NewRouter()
	r.Use(middleware.TrustedSubnetMiddleware(cfg.TrustedSubnet))
	r.Get("/api/internal/stats", handlers.HandleGetStats())

	tests := []struct {
		name           string
		clientIP       string
		expectedStatus int
		expectedURLs   int
		expectedUsers  int
	}{
		{
			name:           "Valid IP in trusted subnet",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
			expectedURLs:   4,
			expectedUsers:  3,
		},
		{
			name:           "IP not in trusted subnet",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedURLs:   0,
			expectedUsers:  0,
		},
		{
			name:           "Missing X-Real-IP header",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
			expectedURLs:   0,
			expectedUsers:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/internal/stats", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			r.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// If access was granted, check response content
			if tt.expectedStatus == http.StatusOK {
				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected content type %s, got %s", "application/json", contentType)
				}

				// Parse response
				var response handlers.StatsResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				// Check expected values
				if response.URLs != tt.expectedURLs {
					t.Errorf("expected %d URLs, got %d", tt.expectedURLs, response.URLs)
				}

				if response.Users != tt.expectedUsers {
					t.Errorf("expected %d users, got %d", tt.expectedUsers, response.Users)
				}
			}
		})
	}
}

func TestStatsEndpointEmptyTrustedSubnet(t *testing.T) {
	// Create test configuration with empty trusted subnet
	cfg := &config.Config{
		TrustedSubnet: "", // Empty trusted subnet should deny all access
	}

	// Create test storage
	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	// Create router
	r := chi.NewRouter()
	r.Use(middleware.TrustedSubnetMiddleware(cfg.TrustedSubnet))
	r.Get("/api/internal/stats", handlers.HandleGetStats())

	// Test with valid IP - should still be denied
	req := httptest.NewRequest("GET", "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should return 403 Forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}
