// Package grpc provides gRPC server implementation for the URL shortening service.
package grpc

import (
	"context"

	"github.com/achufistov/shortygopher.git/internal/app/service"
	"github.com/achufistov/shortygopher.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the ShortenerService gRPC interface.
type Server struct {
	proto.UnimplementedShortenerServiceServer
	service *service.Service
}

// NewServer creates a new gRPC server instance.
func NewServer(service *service.Service) *Server {
	return &Server{
		service: service,
	}
}

// ShortenURL handles gRPC requests to shorten URLs.
func (s *Server) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	serviceReq := service.ShortenURLRequest{
		OriginalURL: req.OriginalUrl,
		UserID:      userID,
	}

	resp, err := s.service.ShortenURL(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	return &proto.ShortenURLResponse{
		ShortUrl:     resp.ShortURL,
		AlreadyExists: resp.AlreadyExists,
	}, nil
}

// GetURL handles gRPC requests to retrieve original URLs.
func (s *Server) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	serviceReq := service.GetURLRequest{
		ShortID: req.ShortId,
	}

	resp, err := s.service.GetURL(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get URL: %v", err)
	}

	if !resp.Exists {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if resp.Deleted {
		return nil, status.Error(codes.Unavailable, "URL has been deleted")
	}

	return &proto.GetURLResponse{
		OriginalUrl: resp.OriginalURL,
		Deleted:     resp.Deleted,
	}, nil
}

// ShortenURLBatch handles gRPC requests to shorten multiple URLs.
func (s *Server) ShortenURLBatch(ctx context.Context, req *proto.ShortenURLBatchRequest) (*proto.ShortenURLBatchResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	serviceURLs := make([]service.BatchRequest, len(req.Urls))
	for i, url := range req.Urls {
		serviceURLs[i] = service.BatchRequest{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	serviceReq := service.ShortenURLBatchRequest{
		URLs:   serviceURLs,
		UserID: userID,
	}

	resp, err := s.service.ShortenURLBatch(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to shorten URLs batch: %v", err)
	}

	protoURLs := make([]*proto.BatchResponse, len(resp.URLs))
	for i, url := range resp.URLs {
		protoURLs[i] = &proto.BatchResponse{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		}
	}

	return &proto.ShortenURLBatchResponse{
		Urls: protoURLs,
	}, nil
}

// GetUserURLs handles gRPC requests to get user's URLs.
func (s *Server) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	serviceReq := service.GetUserURLsRequest{
		UserID: userID,
	}

	resp, err := s.service.GetUserURLs(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	protoURLs := make([]*proto.UserURL, len(resp.URLs))
	for i, url := range resp.URLs {
		protoURLs[i] = &proto.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return &proto.GetUserURLsResponse{
		Urls: protoURLs,
	}, nil
}

// DeleteUserURLs handles gRPC requests to delete user's URLs.
func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	serviceReq := service.DeleteUserURLsRequest{
		ShortURLs: req.ShortUrls,
		UserID:    userID,
	}

	resp, err := s.service.DeleteUserURLs(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete URLs: %v", err)
	}

	return &proto.DeleteUserURLsResponse{
		Success: resp.Success,
	}, nil
}

// Ping handles gRPC health check requests.
func (s *Server) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	serviceReq := service.PingRequest{}
	resp, err := s.service.Ping(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage ping failed: %v", err)
	}

	return &proto.PingResponse{
		Ok: resp.OK,
	}, nil
}

// GetStats handles gRPC requests for service statistics.
func (s *Server) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	// Note: Access control for trusted subnet should be implemented via gRPC middleware
	serviceReq := service.GetStatsRequest{}
	resp, err := s.service.GetStats(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	return &proto.GetStatsResponse{
		Urls:  int32(resp.URLs),
		Users: int32(resp.Users),
	}, nil
}

 