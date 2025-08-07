package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/achufistov/shortygopher.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Get gRPC address from environment or use default
	grpcAddr := "localhost:9090"
	if envAddr := os.Getenv("GRPC_ADDRESS"); envAddr != "" {
		grpcAddr = envAddr
	}

	// Connect to gRPC server
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewShortenerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Add user ID to metadata
	md := metadata.Pairs("user-id", "test-user-123")
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Test Ping
	fmt.Println("Testing Ping...")
	pingResp, err := client.Ping(ctx, &proto.PingRequest{})
	if err != nil {
		log.Printf("Ping failed: %v", err)
	} else {
		fmt.Printf("Ping response: %v\n", pingResp.Ok)
	}

	// Test ShortenURL
	fmt.Println("\nTesting ShortenURL...")
	shortenResp, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
		OriginalUrl: "https://example.com/very/long/url",
	})
	if err != nil {
		log.Printf("ShortenURL failed: %v", err)
	} else {
		fmt.Printf("Shortened URL: %s (already exists: %v)\n", shortenResp.ShortUrl, shortenResp.AlreadyExists)
	}

	// Test ShortenURLBatch
	fmt.Println("\nTesting ShortenURLBatch...")
	batchResp, err := client.ShortenURLBatch(ctx, &proto.ShortenURLBatchRequest{
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
		log.Printf("ShortenURLBatch failed: %v", err)
	} else {
		fmt.Println("Batch shortened URLs:")
		for _, url := range batchResp.Urls {
			fmt.Printf("  %s: %s\n", url.CorrelationId, url.ShortUrl)
		}
	}

	// Test GetUserURLs
	fmt.Println("\nTesting GetUserURLs...")
	userURLsResp, err := client.GetUserURLs(ctx, &proto.GetUserURLsRequest{})
	if err != nil {
		log.Printf("GetUserURLs failed: %v", err)
	} else {
		fmt.Printf("User URLs count: %d\n", len(userURLsResp.Urls))
		for _, url := range userURLsResp.Urls {
			fmt.Printf("  %s -> %s\n", url.ShortUrl, url.OriginalUrl)
		}
	}

	// Test GetStats (without user ID in metadata)
	fmt.Println("\nTesting GetStats...")
	statsCtx := context.Background()
	statsResp, err := client.GetStats(statsCtx, &proto.GetStatsRequest{})
	if err != nil {
		log.Printf("GetStats failed: %v", err)
	} else {
		fmt.Printf("Stats: URLs=%d, Users=%d\n", statsResp.Urls, statsResp.Users)
	}

	fmt.Println("\nAll tests completed!")
}
