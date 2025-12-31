package tm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Private key to ensure only this package can access the Tx in the context
type txKey struct{}

// TransactionManager interface
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type pgxTransactionManager struct {
	db *pgxpool.Pool
}

func NewTransactionManager(db *pgxpool.Pool) TransactionManager {
	return &pgxTransactionManager{db: db}
}

// RunInTransaction manages the transaction lifecycle
func (tm *pgxTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 1. Start Transaction
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Setup Safety Net (Defer Rollback)
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}

		tx.Rollback(ctx)
	}()

	// Inject Transaction into Context
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// Run Business Logic
	if err := fn(txCtx); err != nil {
		return err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTx extracts the transaction from the context (Used by Repositories)
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
