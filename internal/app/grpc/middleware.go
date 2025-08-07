// Package grpc provides gRPC server implementation for the URL shortening service.
package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserIDKey is the context key for user ID
type UserIDKey struct{}

// AuthInterceptor provides authentication middleware for gRPC
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for Ping and GetStats methods
		if info.FullMethod == "/shortener.ShortenerService/Ping" || 
		   info.FullMethod == "/shortener.ShortenerService/GetStats" {
			return handler(ctx, req)
		}

		// Extract user ID from metadata (for now, we'll use a simple approach)
		// In a real implementation, you would extract JWT token and validate it
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("user-id")
		if len(userIDs) == 0 {
			// For now, generate a default user ID
			// In production, you would validate JWT tokens here
			userID := "default-user"
			newCtx := context.WithValue(ctx, UserIDKey{}, userID)
			return handler(newCtx, req)
		}

		userID := strings.TrimSpace(userIDs[0])
		if userID == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid user ID")
		}

		newCtx := context.WithValue(ctx, UserIDKey{}, userID)
		return handler(newCtx, req)
	}
}

// getUserIDFromContext extracts user ID from gRPC context
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey{}).(string); ok {
		return userID
	}
	return ""
} 