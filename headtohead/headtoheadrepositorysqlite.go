package headtohead

import (
	"context"
	"log/slog"
	"muttley/sqlite/entities"
)

type HeadToHeadRepositorySqlite struct {
	queries *entities.Queries
}

func NewHeadToHeadRepository(queries *entities.Queries) *HeadToHeadRepositorySqlite {
	return &HeadToHeadRepositorySqlite{
		queries: queries,
	}
}

func (r *HeadToHeadRepositorySqlite) GetHeadToHead(ctx context.Context, userId int64) ([]entities.GetHeadToHeadRow, error) {
	results, err := r.queries.GetHeadToHead(ctx, userId)
	if err != nil {
		slog.Error(err.Error())
	}
	return results, err
}
