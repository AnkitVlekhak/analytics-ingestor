package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	pb "github.com/ankitlekhak/analytics-ingestor/pkg/api/v1"
)

type App struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
	Container  *Container
	Pool       *pgxpool.Pool
}

func NewApp(dsn string, port string) (*App, error) {
	ctx := context.Background()

	// 1. DB CONFIGURATION (Connection Pooling)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// 2. CONNECT TO DB
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}
	log.Println("✅ Database Connection Pool Initialized")

	// 3. INITIALIZE CONTAINER (The Wiring)
	deps := NewContainer(pool)

	// 4. SETUP gRPC SERVER
	grpcServer := grpc.NewServer()

	// Register the service from the container
	pb.RegisterAnalyticsServiceServer(grpcServer, deps.IngestorService)

	// 5. START LISTENER
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	return &App{
		GRPCServer: grpcServer,
		Listener:   lis,
		Container:  deps,
		Pool:       pool,
	}, nil
}

func (a *App) Run(ctx context.Context) error {

	a.Container.IngestorService.StartBatchProcessing(ctx)

	log.Printf("🚀 gRPC Server listening on %s", a.Listener.Addr())
	return a.GRPCServer.Serve(a.Listener)
}

func (a *App) Shutdown() {
	log.Println("🛑 Shutting down server...")
	a.ObjectShutdown()
}

func (a *App) ObjectShutdown() {
	if a.GRPCServer != nil {
		a.GRPCServer.GracefulStop()
	}
	if a.Pool != nil {
		a.Pool.Close()
		log.Println("🛑 Database Connection Pool Closed")
	}
}
