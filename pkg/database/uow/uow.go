package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/order/pkg/database"
)

var (
	// ErrTransactionAlreadyFinished is returned when trying to rollback a finished transaction
	ErrTransactionAlreadyFinished = errors.New("transaction has already been committed or rolled back")
)

type UnitOfWork interface {
	DBTX() database.DBTX
	Do(ctx context.Context, fn func(ctx context.Context, db database.DBTX) error) error
}

type unitOfWork struct {
	db      *sql.DB
	options *sql.TxOptions
}

// UnitOfWorkOption defines options for creating a UnitOfWork
type UnitOfWorkOption func(*unitOfWork)

// WithIsolationLevel sets the transaction isolation level
func WithIsolationLevel(level sql.IsolationLevel) UnitOfWorkOption {
	return func(u *unitOfWork) {
		if u.options == nil {
			u.options = &sql.TxOptions{}
		}
		u.options.Isolation = level
	}
}

// WithReadOnly sets the transaction as read-only
func WithReadOnly(readOnly bool) UnitOfWorkOption {
	return func(u *unitOfWork) {
		if u.options == nil {
			u.options = &sql.TxOptions{}
		}
		u.options.ReadOnly = readOnly
	}
}

// NewUnitOfWork creates a new UnitOfWork instance
// This instance is safe for concurrent use as it doesn't maintain transaction state
func NewUnitOfWork(db *sql.DB, opts ...UnitOfWorkOption) UnitOfWork {
	u := &unitOfWork{
		db:      db,
		options: nil,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// DBTX returns the underlying database connection
// Note: This returns the connection pool, not an active transaction
// Active transactions should only be accessed within the Do() callback
func (u *unitOfWork) DBTX() database.DBTX {
	return u.db
}

// Do executes the given function within a database transaction
// It handles transaction lifecycle: begin, commit/rollback, and panic recovery
// The transaction is created locally and not stored in the struct to avoid race conditions
func (u *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context, db database.DBTX) error) error {
	// Check if context is already cancelled before starting transaction
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before transaction start: %w", err)
	}

	// Begin transaction with configured options
	tx, err := u.db.BeginTx(ctx, u.options)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Track whether transaction was already finished to avoid double rollback
	var finished bool

	// Ensure transaction is properly closed even in case of panic
	defer func() {
		if p := recover(); p != nil {
			// In case of panic, attempt rollback if transaction wasn't finished
			if !finished {
				_ = tx.Rollback() // Ignore error during panic recovery
			}
			// Re-throw panic after cleanup
			panic(p)
		}
	}()

	// Execute the business logic within the transaction
	if err = fn(ctx, tx); err != nil {
		// Mark as finished before rollback
		finished = true
		if rbErr := rollbackTx(tx); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Attempt to commit the transaction
	finished = true
	if err = tx.Commit(); err != nil {
		// If commit fails, try rollback (some drivers auto-rollback, others don't)
		if rbErr := rollbackTx(tx); rbErr != nil {
			// Only log rollback error if it's not "already finished"
			if !errors.Is(rbErr, ErrTransactionAlreadyFinished) {
				return fmt.Errorf("commit error: %w, rollback error: %v", err, rbErr)
			}
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackTx safely attempts to rollback a transaction
// It handles the case where the transaction might have already been rolled back
func rollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return ErrTransactionAlreadyFinished
	}

	if err := tx.Rollback(); err != nil {
		// Check if error is due to transaction already being finished
		// Different drivers return different errors, but sql.ErrTxDone is standard
		if errors.Is(err, sql.ErrTxDone) {
			return ErrTransactionAlreadyFinished
		}
		return err
	}

	return nil
}
