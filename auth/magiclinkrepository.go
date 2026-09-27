package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrMagicLinkInvalid = errors.New("magic link is invalid or expired")

type MagicLinkRepository interface {
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	InvalidateUnusedForUser(ctx context.Context, userID int64) error
	Consume(ctx context.Context, tokenHash string) (int64, error)
}

type MagicLinkRepositorySqlite struct {
	db *sql.DB
}

func NewMagicLinkRepository(db *sql.DB) *MagicLinkRepositorySqlite {
	return &MagicLinkRepositorySqlite{db: db}
}

func (r *MagicLinkRepositorySqlite) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO magic_link (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *MagicLinkRepositorySqlite) InvalidateUnusedForUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE magic_link SET used_at = ? WHERE user_id = ? AND used_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339), userID,
	)
	return err
}

func (r *MagicLinkRepositorySqlite) Consume(ctx context.Context, tokenHash string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var userID int64
	var expiresAt string
	var usedAt sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used_at FROM magic_link WHERE token_hash = ?`,
		tokenHash,
	).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrMagicLinkInvalid
		}
		return 0, err
	}
	if usedAt.Valid {
		return 0, ErrMagicLinkInvalid
	}
	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return 0, err
	}
	if time.Now().UTC().After(expires) {
		return 0, ErrMagicLinkInvalid
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE magic_link SET used_at = ? WHERE token_hash = ? AND used_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339), tokenHash,
	)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}
