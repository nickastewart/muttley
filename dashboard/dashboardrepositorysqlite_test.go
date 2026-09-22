package dashboard_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"muttley/dashboard"
	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestGetDashboard(t *testing.T) {
	db := testdb.Open(t)
	repo := dashboard.NewDashboardRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")

	// Positions 1, 1, 2, 4, 5: two wins, three podiums, average 2.6.
	for i, position := range []int64{1, 1, 2, 4, 5} {
		seedRace(t, db, ada.ID, "Whilton Mill", fmt.Sprintf("2024-04-%02d", i+1), position)
	}

	stats, err := repo.GetDashboard(context.Background(), ada.ID)
	if err != nil {
		t.Fatalf("get dashboard: %v", err)
	}
	if stats.Totalraces != 5 || stats.Totalwins != 2 || stats.Winrate != 40 || stats.Totalpodiums != 3 || stats.Podiumrate != 60 || stats.Bestposition != 1 || stats.Avgposition != 3 {
		t.Fatalf("dashboard = %+v", stats)
	}
}

func TestGetBestTrack(t *testing.T) {
	db := testdb.Open(t)
	repo := dashboard.NewDashboardRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")

	for i, position := range []int64{1, 2, 1} {
		seedRace(t, db, ada.ID, "Fast Track", fmt.Sprintf("2024-05-%02d", i+1), position)
	}
	for i, position := range []int64{6, 6, 9} {
		seedRace(t, db, ada.ID, "Slow Track", fmt.Sprintf("2024-06-%02d", i+1), position)
	}

	best, err := repo.GetBestTrack(context.Background(), ada.ID)
	if err != nil {
		t.Fatalf("get best track: %v", err)
	}
	// Fast Track average is 4/3, which ceilings to 2. Slow Track averages 7.
	if best.Name != "Fast Track" || best.Avgposition != 2 {
		t.Fatalf("best track = %+v", best)
	}
}

func TestGetBestTrackRequiresThreeRaces(t *testing.T) {
	db := testdb.Open(t)
	repo := dashboard.NewDashboardRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	seedRace(t, db, ada.ID, "Fast Track", "2024-05-01", 1)
	seedRace(t, db, ada.ID, "Fast Track", "2024-05-02", 1)

	_, err := repo.GetBestTrack(context.Background(), ada.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func TestGetLocationStats(t *testing.T) {
	db := testdb.Open(t)
	repo := dashboard.NewDashboardRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	other := createUser(t, db, "Grace", "grace@example.com", "grace")

	counts := []struct {
		name  string
		count int
	}{
		{"Alpha", 4},
		{"Bravo", 3},
		{"Charlie", 2},
		{"Delta", 1},
	}
	day := 1
	for _, track := range counts {
		for range track.count {
			seedRace(t, db, ada.ID, track.name, fmt.Sprintf("2024-03-%02d", day), 1)
			day++
		}
	}

	stats, err := repo.GetLocationStats(context.Background(), ada.ID)
	if err != nil {
		t.Fatalf("get location stats: %v", err)
	}
	if len(stats) != 3 {
		t.Fatalf("got %d locations, want the top 3: %+v", len(stats), stats)
	}
	if stats[0].Name != "Alpha" || stats[0].Count != 4 || stats[1].Name != "Bravo" || stats[1].Count != 3 || stats[2].Name != "Charlie" || stats[2].Count != 2 {
		t.Fatalf("location stats = %+v", stats)
	}

	empty, err := repo.GetLocationStats(context.Background(), other.ID)
	if err != nil {
		t.Fatalf("get empty location stats: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty stats = %+v", empty)
	}
}

func TestGetRecentPositions(t *testing.T) {
	db := testdb.Open(t)
	repo := dashboard.NewDashboardRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	other := createUser(t, db, "Grace", "grace@example.com", "grace")

	for day := 1; day <= 21; day++ {
		seedRace(t, db, ada.ID, "Whilton Mill", fmt.Sprintf("2024-01-%02d", day), int64(day))
	}

	positions, err := repo.GetRecentPositions(context.Background(), ada.ID)
	if err != nil {
		t.Fatalf("get recent positions: %v", err)
	}
	if len(positions) != 20 {
		t.Fatalf("got %d positions, want 20", len(positions))
	}
	if positions[0] != 21 || positions[19] != 2 {
		t.Fatalf("positions = %v, want newest 21 down to 2", positions)
	}

	none, err := repo.GetRecentPositions(context.Background(), other.ID)
	if err != nil {
		t.Fatalf("get positions with no races: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("positions = %v, want none", none)
	}
}

func createUser(t *testing.T, db *sql.DB, firstName, email, profileID string) entities.GetUserByIdRow {
	t.Helper()
	repo := user.NewUserRepository(db)
	ctx := context.Background()
	if _, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    "Racer",
		Email:       email,
		Password:    "secret",
		ProfileID:   profileID,
		DisplayName: firstName,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	byEmail, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	byID, err := repo.GetUserById(ctx, byEmail.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	return byID
}

func seedRace(t *testing.T, db *sql.DB, userID int64, track, date string, position int64) {
	t.Helper()
	ctx := context.Background()
	locations := location.NewLocationRepository(entities.New(db))
	loc, err := locations.GetLocationByName(ctx, track)
	if errors.Is(err, sql.ErrNoRows) {
		loc, err = locations.CreateLocation(ctx, track)
	}
	if err != nil {
		t.Fatalf("location %s: %v", track, err)
	}

	race, err := event.NewEventRepository(entities.New(db)).CreateEvent(ctx, entities.CreateEventParams{
		LocationID:   loc.ID,
		Type:         "Rental",
		Date:         date,
		TotalDrivers: 10,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := eventresult.NewEventResultRepository(entities.New(db)).CreateEventResult(ctx, entities.CreateEventResultParams{
		EventID:        race.ID,
		UserID:         userID,
		BestLapTime:    45000,
		AverageLapTime: 47000,
		Position:       position,
		NumberOfLaps:   12,
	}); err != nil {
		t.Fatalf("create result: %v", err)
	}
}
