package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/service"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/achufistov/shortygopher.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGRPCIntegration(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Address:     "localhost:8084",
		GRPCAddress: "localhost:9094",
		BaseURL:     "http://localhost:8084",
		FileStorage: "",
		DatabaseDSN: "",
		SecretKey:   "test-secret-key",
	}

	// Initialize storage and service
	storageInstance := storage.NewURLStorage()
	serviceInstance := service.NewService(storageInstance, cfg)

	// Start gRPC server
	grpcServer := grpc.NewServer()
	grpcService := NewTestGRPCServer(serviceInstance)
	proto.RegisterShortenerServiceServer(grpcServer, grpcService)

	// Start server in goroutine
	go func() {
		listener, err := net.Listen("tcp", cfg.GRPCAddress)
		if err != nil {
			t.Fatalf("Failed to listen: %v", err)
		}
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(time.Second)

	// Connect to gRPC server
	conn, err := grpc.Dial(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewShortenerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Add user ID to metadata
	md := metadata.Pairs("user-id", "test-user-123")
	ctx = metadata.NewOutgoingContext(ctx, md)

	t.Run("TestPing", func(t *testing.T) {
		resp, err := client.Ping(ctx, &proto.PingRequest{})
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
		if !resp.Ok {
			t.Error("Expected ping to be successful")
		}
	})

	t.Run("TestShortenURL", func(t *testing.T) {
		resp, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
			OriginalUrl: "https://example.com/test",
		})
		if err != nil {
			t.Fatalf("ShortenURL failed: %v", err)
		}
		if resp.ShortUrl == "" {
			t.Error("Expected short URL to be generated")
		}
		if resp.AlreadyExists {
			t.Error("Expected URL to be new")
		}
	})

	t.Run("TestShortenURLBatch", func(t *testing.T) {
		resp, err := client.ShortenURLBatch(ctx, &proto.ShortenURLBatchRequest{
			Urls: []*proto.BatchRequest{
				{
					CorrelationId: "1",
					OriginalUrl:   "https://google.com",
				},
				{
					CorrelationId: "2",
					OriginalUrl:   "https://github.com",
				},
			},
		})
		if err != nil {
			t.Fatalf("ShortenURLBatch failed: %v", err)
		}
		if len(resp.Urls) != 2 {
			t.Errorf("Expected 2 URLs, got %d", len(resp.Urls))
		}
	})

	t.Run("TestGetUserURLs", func(t *testing.T) {
		resp, err := client.GetUserURLs(ctx, &proto.GetUserURLsRequest{})
		if err != nil {
			t.Fatalf("GetUserURLs failed: %v", err)
		}
		if len(resp.Urls) < 3 {
			t.Errorf("Expected at least 3 URLs, got %d", len(resp.Urls))
		}
	})

	t.Run("TestGetStats", func(t *testing.T) {
		// Test without user ID in metadata
		statsCtx := context.Background()
		resp, err := client.GetStats(statsCtx, &proto.GetStatsRequest{})
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}
		if resp.Urls < 3 {
			t.Errorf("Expected at least 3 URLs in stats, got %d", resp.Urls)
		}
		if resp.Users < 1 {
			t.Errorf("Expected at least 1 user in stats, got %d", resp.Users)
		}
	})

	// Graceful shutdown
	grpcServer.GracefulStop()
}

// TestGRPCServer is a simplified version for testing
type TestGRPCServer struct {
	proto.UnimplementedShortenerServiceServer
	service *service.Service
}

func NewTestGRPCServer(service *service.Service) *TestGRPCServer {
	return &TestGRPCServer{
		service: service,
	}
}

func (s *TestGRPCServer) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	serviceReq := service.PingRequest{}
	resp, err := s.service.Ping(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	return &proto.PingResponse{Ok: resp.OK}, nil
}

func (s *TestGRPCServer) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID := "test-user-123" // Simplified for testing
	serviceReq := service.ShortenURLRequest{
		OriginalURL: req.OriginalUrl,
		UserID:      userID,
	}
	resp, err := s.service.ShortenURL(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	return &proto.ShortenURLResponse{
		ShortUrl:     resp.ShortURL,
		AlreadyExists: resp.AlreadyExists,
	}, nil
}

func (s *TestGRPCServer) ShortenURLBatch(ctx context.Context, req *proto.ShortenURLBatchRequest) (*proto.ShortenURLBatchResponse, error) {
	userID := "test-user-123" // Simplified for testing
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
		return nil, err
	}
	protoURLs := make([]*proto.BatchResponse, len(resp.URLs))
	for i, url := range resp.URLs {
		protoURLs[i] = &proto.BatchResponse{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		}
	}
	return &proto.ShortenURLBatchResponse{Urls: protoURLs}, nil
}

func (s *TestGRPCServer) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	userID := "test-user-123" // Simplified for testing
	serviceReq := service.GetUserURLsRequest{
		UserID: userID,
	}
	resp, err := s.service.GetUserURLs(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	protoURLs := make([]*proto.UserURL, len(resp.URLs))
	for i, url := range resp.URLs {
		protoURLs[i] = &proto.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}
	return &proto.GetUserURLsResponse{Urls: protoURLs}, nil
}

func (s *TestGRPCServer) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	serviceReq := service.GetStatsRequest{}
	resp, err := s.service.GetStats(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	return &proto.GetStatsResponse{
		Urls:  int32(resp.URLs),
		Users: int32(resp.Users),
	}, nil
} 