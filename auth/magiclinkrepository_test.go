package auth

import (
	"context"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"testing"
	"time"
)

func TestMagicLinkConsume(t *testing.T) {
	db := testdb.Open(t)
	queries := entities.New(db)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
		Password:    "hashed",
		ProfileID:   "ada-lovelace-1",
		DisplayName: "Ada",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// CreateUser RETURNING does not include id; look the user up.
	found, err := queries.GetUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	_ = user

	repo := NewMagicLinkRepository(db)
	raw, hash, err := generateMagicToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if hashMagicToken(raw) != hash {
		t.Fatal("hash mismatch")
	}

	if err := repo.Create(ctx, found.ID, hash, time.Now().Add(15*time.Minute)); err != nil {
		t.Fatalf("create link: %v", err)
	}

	userID, err := repo.Consume(ctx, hash)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if userID != found.ID {
		t.Fatalf("got user %d want %d", userID, found.ID)
	}

	if _, err := repo.Consume(ctx, hash); err != ErrMagicLinkInvalid {
		t.Fatalf("expected reuse to fail, got %v", err)
	}
}

func TestMagicLinkExpired(t *testing.T) {
	db := testdb.Open(t)
	queries := entities.New(db)
	ctx := context.Background()

	if _, err := queries.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
		Password:    "hashed",
		ProfileID:   "ada-lovelace-1",
		DisplayName: "Ada",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	found, err := queries.GetUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	repo := NewMagicLinkRepository(db)
	_, hash, err := generateMagicToken()
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, found.ID, hash, time.Now().Add(-1*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Consume(ctx, hash); err != ErrMagicLinkInvalid {
		t.Fatalf("expected expired token to fail, got %v", err)
	}
}
