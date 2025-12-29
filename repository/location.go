package repository

import (
	"context"
	"log"
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
		log.Println(err)
	}

	return location, err
}

func (r *LocationRepositorySqlite) GetLocationByName(ctx context.Context, name string) (entities.Location, error) {
	location, err := r.queries.GetLocationByName(ctx, name)

	if err != nil {
		log.Println(err)
	}

	return location, err
}
