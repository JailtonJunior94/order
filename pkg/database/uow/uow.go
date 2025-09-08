package uow

import (
	"context"
	"database/sql"
)

type UnitOfWork interface {
	Executor() Executor
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type unitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) *unitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (u *unitOfWork) Executor() Executor {
	if u.tx != nil {
		return u.tx
	}
	return u.db
}
