package location

import (
	"context"
	"log/slog"
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

func (r *LocationRepositorySqlite) CreateLocation(ctx context.Context, name string) (entities.Location, error) {
	location, err := r.queries.CreateLocation(ctx, name)

	if err != nil {
		slog.Error(err.Error())
	}

	return location, err
}

func (r *LocationRepositorySqlite) GetLocationByName(ctx context.Context, name string) (entities.Location, error) {
	location, err := r.queries.GetLocationByName(ctx, name)

	if err != nil {
		slog.Error(err.Error())
	}

	return location, err
}
