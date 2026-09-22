package sqlite

import (
	"context"
	"database/sql"
	"muttley/sqlite/entities"
)

type txKey struct{}

type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) Within(ctx context.Context, fn func(context.Context) error) error {
	if inTx(ctx) {
		return fn(ctx)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func Queries(ctx context.Context, queries *entities.Queries) *entities.Queries {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok || tx == nil {
		return queries
	}
	return queries.WithTx(tx)
}

func inTx(ctx context.Context) bool {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return ok && tx != nil
}
