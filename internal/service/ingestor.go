package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ankitlekhak/analytics-ingestor/internal/repository"
	pb "github.com/ankitlekhak/analytics-ingestor/pkg/api/v1"
	"github.com/ankitlekhak/analytics-ingestor/pkg/tm"
)

const (
	BatchSize     = 5               // Flush when we have 100 items
	FlushInterval = 1 * time.Second // Flush every 1 second
	BufferLimit   = 10              // Max channel capacity
)

type IngestorService struct {
	pb.UnimplementedAnalyticsServiceServer
	metricChan         chan *pb.LogMetricsRequest
	metricRepo         repository.MetricRepository
	transactionManager tm.TransactionManager
}

func NewIngestorService(metricRepo repository.MetricRepository, transactionManager tm.TransactionManager) *IngestorService {
	s := &IngestorService{
		metricChan:         make(chan *pb.LogMetricsRequest, BufferLimit),
		metricRepo:         metricRepo,
		transactionManager: transactionManager,
	}

	return s
}

func (s *IngestorService) StartBatchProcessing(ctx context.Context) {
	log.Println("👷 Starting Ingestor Background Worker...")
	go s.startBatchWorker(ctx)
}

func (s *IngestorService) startBatchWorker(ctx context.Context) {
	var batch []*pb.LogMetricsRequest
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			if len(batch) > 0 {
				log.Println("🛑 Context cancelled, flushing remaining batch...")
				s.flushBatch(ctx, batch)
			}
			return

		case metric, ok := <-s.metricChan:
			if !ok {

				if len(batch) > 0 {
					s.flushBatch(ctx, batch)
				}
				return
			}

			batch = append(batch, metric)
			if len(batch) >= BatchSize {
				s.flushBatch(ctx, batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(ctx, batch)
				batch = nil
			}
		}
	}
}

func (s *IngestorService) flushBatch(parentCtx context.Context, batch []*pb.LogMetricsRequest) {
	log.Printf("💾 Flushing batch of %d metrics", len(batch))

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	err := s.transactionManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		return s.metricRepo.BulkInsertMetrics(txCtx, batch)
	})

	if err != nil {
		log.Printf("❌ Transaction failed: %v", err)
	} else {
		fmt.Printf("💾 Batch committed successfully\n")
	}

}

func (s *IngestorService) LogMetric(ctx context.Context, req *pb.LogMetricsRequest) (*pb.LogMetricsResponse, error) {
	select {
	case s.metricChan <- req:
		return &pb.LogMetricsResponse{Success: true}, nil
	default:
		log.Println("⚠️ Buffer full, dropping metric")
		return &pb.LogMetricsResponse{Success: false}, nil
	}
}
