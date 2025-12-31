package app

import (
	"github.com/ankitlekhak/analytics-ingestor/internal/repository"
	"github.com/ankitlekhak/analytics-ingestor/internal/service"
	"github.com/ankitlekhak/analytics-ingestor/pkg/tm"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	IngestorService *service.IngestorService
}

func NewContainer(db *pgxpool.Pool) *Container {
	transactionManager := tm.NewTransactionManager(db)

	metricRepo := repository.NewTimescaleMetricRepository(db)

	ingestorService := service.NewIngestorService(metricRepo, transactionManager)
	return &Container{
		IngestorService: ingestorService,
	}
}
