package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "Empty trusted subnet denies all access",
			trustedSubnet:  "",
			clientIP:       "192.168.1.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid client IP",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Client IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Client IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Client IP at subnet boundary",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.1.255",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Client IP outside subnet boundary",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.2.1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := TrustedSubnetMiddleware(tt.trustedSubnet)

			// Create a simple handler that returns 200 OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply middleware and call handler
			middleware(handler).ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestTrustedSubnetMiddlewareInvalidCIDR(t *testing.T) {
	// Test with invalid CIDR notation
	middleware := TrustedSubnetMiddleware("invalid-cidr")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")

	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Should return 500 Internal Server Error for invalid CIDR
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}
