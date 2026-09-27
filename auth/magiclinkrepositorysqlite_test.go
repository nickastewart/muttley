package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"muttley/auth"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
)

func TestMagicLinkIsSingleUseAndExpires(t *testing.T) {
	db := testdb.Open(t)
	repo := auth.NewMagicLinkRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	created, err := repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "ada@example.com",
		TokenHash: "hash",
		Purpose:   "login",
		ExpiresAt: now.Add(time.Minute).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create magic link: %v", err)
	}

	got, err := repo.GetActiveByTokenHash(ctx, entities.GetActiveMagicLinkByTokenHashParams{
		TokenHash: "hash",
		Now:       now.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("get active magic link: %v", err)
	}
	if got.ID != created.ID || got.Email != "ada@example.com" || got.Purpose != "login" {
		t.Fatalf("active link = %+v", got)
	}

	consumed, err := repo.Consume(ctx, got.ID)
	if err != nil {
		t.Fatalf("consume magic link: %v", err)
	}
	if consumed != 1 {
		t.Fatalf("consumed = %d, want 1", consumed)
	}
	consumed, err = repo.Consume(ctx, got.ID)
	if err != nil {
		t.Fatalf("consume magic link again: %v", err)
	}
	if consumed != 0 {
		t.Fatalf("consumed again = %d, want 0", consumed)
	}
	_, err = repo.GetActiveByTokenHash(ctx, entities.GetActiveMagicLinkByTokenHashParams{
		TokenHash: "hash",
		Now:       now.Format(time.RFC3339),
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("used link error = %v, want sql.ErrNoRows", err)
	}

	_, err = repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "ada@example.com",
		TokenHash: "expired",
		Purpose:   "signup",
		ExpiresAt: now.Add(-time.Minute).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create expired magic link: %v", err)
	}
	_, err = repo.GetActiveByTokenHash(ctx, entities.GetActiveMagicLinkByTokenHashParams{
		TokenHash: "expired",
		Now:       now.Format(time.RFC3339),
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expired link error = %v, want sql.ErrNoRows", err)
	}

	fresh, err := repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "ada@example.com",
		TokenHash: "fresh",
		Purpose:   "login",
		ExpiresAt: now.Add(time.Minute).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create fresh magic link: %v", err)
	}
	if err := repo.InvalidateUnused(ctx, "ada@example.com"); err != nil {
		t.Fatalf("invalidate unused: %v", err)
	}
	_, err = repo.GetActiveByTokenHash(ctx, entities.GetActiveMagicLinkByTokenHashParams{
		TokenHash: "fresh",
		Now:       now.Format(time.RFC3339),
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("invalidated link %d error = %v, want sql.ErrNoRows", fresh.ID, err)
	}
}
