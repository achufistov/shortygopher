// Package middleware provides HTTP middleware functions for the URL shortening service.
package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware creates middleware that checks if the client's IP address
// is within the trusted subnet defined in CIDR notation.
// If trustedSubnet is empty, access is denied for all requests.
// The client IP is extracted from the X-Real-IP header.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If trusted subnet is empty, deny access
			if trustedSubnet == "" {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Get client IP from X-Real-IP header
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			// Parse the trusted subnet
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			// Parse the client IP
			ip := net.ParseIP(clientIP)
			if ip == nil {
				http.Error(w, "Invalid client IP address", http.StatusForbidden)
				return
			}

			// Check if client IP is within the trusted subnet
			if !subnet.Contains(ip) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
