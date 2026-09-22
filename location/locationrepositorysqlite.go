package location

import (
	"context"
	"log/slog"
	"muttley/sqlite"
	"muttley/sqlite/entities"
)

type LocationRepositorySqlite struct {
	queries *entities.Queries
}

func NewLocationRepository(queries *entities.Queries) *LocationRepositorySqlite {
	return &LocationRepositorySqlite{
		queries: queries,
	}
}

func (r *LocationRepositorySqlite) q(ctx context.Context) *entities.Queries {
	return sqlite.Queries(ctx, r.queries)
}

func (r *LocationRepositorySqlite) CreateLocation(ctx context.Context, name string) (entities.Location, error) {
	location, err := r.q(ctx).CreateLocation(ctx, name)

	if err != nil {
		slog.Error(err.Error())
	}

	return location, err
}

func (r *LocationRepositorySqlite) GetLocationByName(ctx context.Context, name string) (entities.Location, error) {
	location, err := r.q(ctx).GetLocationByName(ctx, name)

	if err != nil {
		slog.Error(err.Error())
	}

	return location, err
}
