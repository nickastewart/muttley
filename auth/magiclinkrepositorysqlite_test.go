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

func TestMagicLinkTokenIsSavedAndReadFromTheDatabase(t *testing.T) {
	db := testdb.Open(t)
	repo := auth.NewMagicLinkRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	expires := now.Add(time.Minute).Format(time.RFC3339)

	if _, err := repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "ada@example.com",
		TokenHash: "hash-one",
		Purpose:   "login",
		ExpiresAt: expires,
	}); err != nil {
		t.Fatalf("create first link: %v", err)
	}
	if err := repo.SaveToken(ctx, "ada@example.com", "first"); err != nil {
		t.Fatalf("save first token: %v", err)
	}
	if err := repo.InvalidateUnused(ctx, "ada@example.com"); err != nil {
		t.Fatalf("invalidate first link: %v", err)
	}
	if _, err := repo.ActiveToken(ctx, "ada@example.com"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("used token error = %v, want sql.ErrNoRows", err)
	}

	if _, err := repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "ada@example.com",
		TokenHash: "hash-two",
		Purpose:   "login",
		ExpiresAt: expires,
	}); err != nil {
		t.Fatalf("create second link: %v", err)
	}
	if _, err := repo.Create(ctx, entities.CreateMagicLinkParams{
		Email:     "grace@example.com",
		TokenHash: "hash-grace",
		Purpose:   "login",
		ExpiresAt: expires,
	}); err != nil {
		t.Fatalf("create grace link: %v", err)
	}
	if err := repo.SaveToken(ctx, "ada@example.com", "second"); err != nil {
		t.Fatalf("save second token: %v", err)
	}
	if err := repo.SaveToken(ctx, "grace@example.com", "other"); err != nil {
		t.Fatalf("save grace token: %v", err)
	}

	ada, err := repo.ActiveToken(ctx, "ada@example.com")
	if err != nil || ada != "second" {
		t.Fatalf("ada token = %q err = %v", ada, err)
	}
	grace, err := repo.ActiveToken(ctx, "grace@example.com")
	if err != nil || grace != "other" {
		t.Fatalf("grace token = %q err = %v", grace, err)
	}

	var stored string
	if err := db.QueryRow(`SELECT token FROM magic_link WHERE email = ? AND used_at IS NULL`, "ada@example.com").Scan(&stored); err != nil {
		t.Fatalf("read stored token: %v", err)
	}
	if stored != "second" {
		t.Fatalf("stored token = %q", stored)
	}
}
