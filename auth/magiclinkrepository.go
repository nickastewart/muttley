package auth

import (
	"context"

	"muttley/sqlite/entities"
)

type MagicLinkRepository interface {
	Create(ctx context.Context, params entities.CreateMagicLinkParams) (entities.MagicLink, error)
	InvalidateUnused(ctx context.Context, email string) error
	GetActiveByTokenHash(ctx context.Context, params entities.GetActiveMagicLinkByTokenHashParams) (entities.MagicLink, error)
	Consume(ctx context.Context, id int64) (int64, error)
}
