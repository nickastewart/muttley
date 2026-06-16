package location

import (
	"context"
	"muttley/sqlite/entities"
)

type LocationRepository interface {
	CreateLocation(ctx context.Context, name string) (entities.Location, error)
	GetLocationByName(ctx context.Context, name string) (entities.Location, error)
}
