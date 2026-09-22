package sqlite_test

import (
	"context"
	"errors"
	"testing"

	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
)

func TestWithinCommitsAndRollsBack(t *testing.T) {
	db := testdb.Open(t)
	queries := entities.New(db)
	transactor := sqlite.NewTransactor(db)
	ctx := context.Background()

	err := transactor.Within(ctx, func(ctx context.Context) error {
		if _, err := sqlite.Queries(ctx, queries).CreateLocation(ctx, "Kept"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	err = transactor.Within(ctx, func(ctx context.Context) error {
		if _, err := sqlite.Queries(ctx, queries).CreateLocation(ctx, "Rolled Back"); err != nil {
			return err
		}
		return errors.New("save failed")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}

	kept, err := queries.GetLocationByName(ctx, "Kept")
	if err != nil {
		t.Fatalf("kept location: %v", err)
	}
	if kept.Name != "Kept" {
		t.Fatalf("kept location = %+v", kept)
	}
	if _, err := queries.GetLocationByName(ctx, "Rolled Back"); err == nil {
		t.Fatal("rolled back location was committed")
	}
}

func TestWithinJoinsAnOpenTransaction(t *testing.T) {
	db := testdb.Open(t)
	queries := entities.New(db)
	transactor := sqlite.NewTransactor(db)

	err := transactor.Within(context.Background(), func(ctx context.Context) error {
		return transactor.Within(ctx, func(ctx context.Context) error {
			_, err := sqlite.Queries(ctx, queries).CreateLocation(ctx, "Nested")
			return err
		})
	})
	if err != nil {
		t.Fatalf("nested within: %v", err)
	}

	if _, err := queries.GetLocationByName(context.Background(), "Nested"); err != nil {
		t.Fatalf("nested location: %v", err)
	}
}
