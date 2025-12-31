package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ankitlekhak/analytics-ingestor/pkg/api/v1"
)

func main() {
	// 1. Connect to the Server (No SSL for local dev)
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Did not connect: %v", err)
	}
	defer conn.Close()

	// 2. Initialize the Client
	client := pb.NewAnalyticsServiceClient(conn)

	fmt.Println("🚀 Starting Load Test...")

	// 3. Flood the server with 100 requests
	for i := 0; i < 100; i++ {
		// Create a timeout for each request
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		// Create a dummy metric
		req := &pb.LogMetricsRequest{
			ServiceName: "payment-service",
			MetricName:  "processing_latency",
			Value:       int64(rand.Intn(100)),
			Timestamp:   timestamppb.Now(),
		}

		// FIRE THE REQUEST (gRPC Call)
		resp, err := client.LogMetric(ctx, req)

		if err != nil {
			log.Printf("❌ Request failed: %v", err)
		} else {
			fmt.Printf("✅ Sent Request %d | Server Ack: %v\n", i, resp.Success)
		}

		cancel()                          // Clean up context
		time.Sleep(50 * time.Millisecond) // Slight delay to see the logs
	}
}
