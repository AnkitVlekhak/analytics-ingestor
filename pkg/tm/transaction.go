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

	// 2. Setup Safety Net (Defer Rollback)
	defer func() {
		// If the function panics, we rollback and re-panic
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
		// We attempt rollback here, but if Commit was already called below,
		// pgx detects it safely and does nothing.
		tx.Rollback(ctx)
	}()

	// 3. Inject Transaction into Context
	// This is the "Secret Sauce" - the Repo will find the Tx here
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// 4. Run Business Logic
	if err := fn(txCtx); err != nil {
		// If logic fails, we return error. The defer above triggers Rollback.
		return err
	}

	// 5. Commit
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
