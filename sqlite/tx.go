// Package sqlite shares one database transaction across repository calls.
// A process that writes several rows starts a transaction and passes that
// context into the repositories. Each repository uses Queries so its
// statements join the open transaction instead of committing on their own.
package sqlite

import (
	"context"
	"database/sql"
	"muttley/sqlite/entities"
)

type txKey struct{}

// Transactor begins and commits a transaction around a group of database calls.
type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

// Within runs fn inside a transaction. The context passed to fn carries that
// transaction. If ctx is already inside a transaction, fn joins it and the
// outer caller commits. fn's error rolls the transaction back.
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

// Queries returns queries bound to the transaction in ctx, or queries unchanged
// when ctx has no transaction.
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
