package repository

import (
	"context"
	"fmt"
	"time"

	pb "github.com/ankitlekhak/analytics-ingestor/pkg/api/v1"
	"github.com/ankitlekhak/analytics-ingestor/pkg/tm"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricRepository interface {
	BulkInsertMetrics(ctx context.Context, batch []*pb.LogMetricsRequest) error
}

type timescaleMetricRepository struct {
	db *pgxpool.Pool
}

func NewTimescaleMetricRepository(db *pgxpool.Pool) MetricRepository {
	return &timescaleMetricRepository{db: db}
}

func (t *timescaleMetricRepository) BulkInsertMetrics(ctx context.Context, batch []*pb.LogMetricsRequest) error {
	// 1. Extract the Transaction from Context
	tx, ok := tm.GetTx(ctx)
	if !ok {
		return fmt.Errorf("transaction missing in context: this method must be run inside a transaction")
	}

	query := "INSERT INTO metrics (time, service_name, metric_name, value) VALUES ($1, $2, $3, $4)"

	for _, m := range batch {
		parsedTime, _ := time.Parse(time.RFC3339, m.Timestamp.AsTime().Format(time.RFC3339))
		if parsedTime.IsZero() {
			parsedTime = time.Now()
		}

		// 2. Use the Tx
		_, err := tx.Exec(ctx, query, parsedTime, m.ServiceName, m.MetricName, m.Value)
		if err != nil {
			return err
		}
	}

	return nil
}
