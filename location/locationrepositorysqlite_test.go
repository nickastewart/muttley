package location_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
)

func TestCreateAndGetLocation(t *testing.T) {
	repo := newLocationRepo(t)
	ctx := context.Background()

	created, err := repo.CreateLocation(ctx, "Whilton Mill")
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	if created.ID == 0 || created.Name != "Whilton Mill" {
		t.Fatalf("created location = %+v", created)
	}

	found, err := repo.GetLocationByName(ctx, "Whilton Mill")
	if err != nil {
		t.Fatalf("get location: %v", err)
	}
	if found.ID != created.ID || found.Name != "Whilton Mill" {
		t.Fatalf("found location = %+v, want %+v", found, created)
	}
}

func TestGetMissingLocation(t *testing.T) {
	repo := newLocationRepo(t)

	_, err := repo.GetLocationByName(context.Background(), "Missing Track")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func newLocationRepo(t *testing.T) location.LocationRepository {
	t.Helper()
	return location.NewLocationRepository(entities.New(testdb.Open(t)))
}
