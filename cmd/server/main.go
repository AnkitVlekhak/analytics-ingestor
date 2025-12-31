package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ankitlekhak/analytics-ingestor/internal/app"
)

func main() {
	dsn := "postgres://postgres:password@localhost:5432/analytics_db"
	port := "50051"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	application, err := app.NewApp(dsn, port)
	if err != nil {
		log.Fatalf("❌ Failed to initialize app: %v", err)
	}

	// 2. Run in a goroutine so we can listen for shutdown signals
	go func() {
		if err := application.Run(ctx); err != nil {
			log.Fatalf("❌ Server crashed: %v", err)
		}
	}()

	// 3. Graceful Shutdown (Wait for Ctrl+C)
	<-ctx.Done()

	application.Shutdown()
	log.Println("✅ Server exited properly")
}
