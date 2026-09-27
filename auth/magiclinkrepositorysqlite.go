package auth

import (
	"context"
	"database/sql"
	"time"

	"muttley/sqlite"
	"muttley/sqlite/entities"
)

type MagicLinkRepositorySqlite struct {
	queries *entities.Queries
}

func NewMagicLinkRepository(db *sql.DB) *MagicLinkRepositorySqlite {
	return &MagicLinkRepositorySqlite{queries: entities.New(db)}
}

func (r *MagicLinkRepositorySqlite) q(ctx context.Context) *entities.Queries {
	return sqlite.Queries(ctx, r.queries)
}

func (r *MagicLinkRepositorySqlite) Create(ctx context.Context, params entities.CreateMagicLinkParams) (entities.MagicLink, error) {
	return r.q(ctx).CreateMagicLink(ctx, params)
}

func (r *MagicLinkRepositorySqlite) InvalidateUnused(ctx context.Context, email string) error {
	return r.q(ctx).InvalidateUnusedMagicLinks(ctx, email)
}

func (r *MagicLinkRepositorySqlite) GetActiveByTokenHash(ctx context.Context, params entities.GetActiveMagicLinkByTokenHashParams) (entities.MagicLink, error) {
	return r.q(ctx).GetActiveMagicLinkByTokenHash(ctx, params)
}

func (r *MagicLinkRepositorySqlite) Consume(ctx context.Context, id int64) (int64, error) {
	return r.q(ctx).ConsumeMagicLink(ctx, id)
}

func (r *MagicLinkRepositorySqlite) ActiveToken(ctx context.Context, email string) (string, error) {
	return r.q(ctx).GetActiveMagicLinkTokenByEmail(ctx, entities.GetActiveMagicLinkTokenByEmailParams{
		Email: email,
		Now:   time.Now().UTC().Format(time.RFC3339),
	})
}
