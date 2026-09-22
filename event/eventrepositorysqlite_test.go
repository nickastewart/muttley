package event_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestCreateAndGetEvent(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	events := event.NewEventRepository(entities.New(db))
	track := createLocation(t, db, "Whilton Mill")

	created, err := events.CreateEvent(ctx, entities.CreateEventParams{
		LocationID:   track.ID,
		Type:         "Rental",
		Date:         "2024-06-01",
		TotalDrivers: 12,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if created.ID == 0 || created.LocationID != track.ID || created.Type != "Rental" || created.Date != "2024-06-01" || created.TotalDrivers != 12 {
		t.Fatalf("created event = %+v", created)
	}

	found, err := events.GetEventByLocationAndTypeAndDate(ctx, entities.GetEventByLocationAndTypeAndDateParams{
		LocationID: track.ID,
		Type:       "Rental",
		Date:       "2024-06-01",
	})
	if err != nil {
		t.Fatalf("get event: %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("found event = %+v, want id %d", found, created.ID)
	}
}

func TestGetMissingEvent(t *testing.T) {
	events := event.NewEventRepository(entities.New(testdb.Open(t)))

	_, err := events.GetEventByLocationAndTypeAndDate(context.Background(), entities.GetEventByLocationAndTypeAndDateParams{
		LocationID: 1,
		Type:       "Rental",
		Date:       "2024-06-01",
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func TestGetEventsByUser(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	events := event.NewEventRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	track := createLocation(t, db, "Whilton Mill")

	adaEvent := createEvent(t, db, track.ID, "2024-06-01")
	graceEvent := createEvent(t, db, track.ID, "2024-06-02")
	createResult(t, db, adaEvent.ID, ada.ID, 1)
	createResult(t, db, graceEvent.ID, grace.ID, 3)

	both, err := events.GetEventsByUser(ctx, []int64{ada.ID, grace.ID})
	if err != nil {
		t.Fatalf("get events for both users: %v", err)
	}
	if len(both) != 2 {
		t.Fatalf("got %d events, want 2", len(both))
	}
	byUser := map[int64]entities.GetEventsByUserRow{}
	for _, row := range both {
		byUser[row.EventResult.UserID] = row
	}
	if byUser[ada.ID].Event.Date != "2024-06-01" || byUser[ada.ID].Location.Name != "Whilton Mill" || byUser[ada.ID].EventResult.Position != 1 || byUser[ada.ID].User.Email != "ada@example.com" {
		t.Fatalf("ada event = %+v", byUser[ada.ID])
	}
	if byUser[grace.ID].Event.Date != "2024-06-02" || byUser[grace.ID].EventResult.Position != 3 {
		t.Fatalf("grace event = %+v", byUser[grace.ID])
	}

	onlyAda, err := events.GetEventsByUser(ctx, []int64{ada.ID})
	if err != nil {
		t.Fatalf("get ada events: %v", err)
	}
	if len(onlyAda) != 1 || onlyAda[0].EventResult.UserID != ada.ID {
		t.Fatalf("ada events = %+v", onlyAda)
	}

	none, err := events.GetEventsByUser(ctx, nil)
	if err != nil {
		t.Fatalf("get events for no users: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("got %d events for no users, want 0", len(none))
	}
}

func TestGetRecentEvents(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	events := event.NewEventRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	track := createLocation(t, db, "Whilton Mill")

	for day := 1; day <= 11; day++ {
		race := createEvent(t, db, track.ID, fmt.Sprintf("2024-01-%02d", day))
		createResult(t, db, race.ID, ada.ID, int64(day))
	}

	recent, err := events.GetRecentEvents(ctx, ada.ID)
	if err != nil {
		t.Fatalf("get recent events: %v", err)
	}
	if len(recent) != 10 {
		t.Fatalf("got %d events, want 10", len(recent))
	}
	if recent[0].Event.Date != "2024-01-11" || recent[0].EventResult.Position != 11 || recent[0].Location.Name != "Whilton Mill" {
		t.Fatalf("newest event = %+v", recent[0])
	}
	if recent[9].Event.Date != "2024-01-02" || recent[9].EventResult.Position != 2 {
		t.Fatalf("oldest returned event = %+v", recent[9])
	}
}

func createUser(t *testing.T, db *sql.DB, firstName, email, profileID string) entities.GetUserByIdRow {
	t.Helper()
	repo := user.NewUserRepository(entities.New(db))
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

func createLocation(t *testing.T, db *sql.DB, name string) entities.Location {
	t.Helper()
	loc, err := location.NewLocationRepository(entities.New(db)).CreateLocation(context.Background(), name)
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc
}

func createEvent(t *testing.T, db *sql.DB, locationID int64, date string) entities.Event {
	t.Helper()
	race, err := event.NewEventRepository(entities.New(db)).CreateEvent(context.Background(), entities.CreateEventParams{
		LocationID:   locationID,
		Type:         "Rental",
		Date:         date,
		TotalDrivers: 10,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	return race
}

func createResult(t *testing.T, db *sql.DB, eventID, userID, position int64) {
	t.Helper()
	_, err := eventresult.NewEventResultRepository(entities.New(db)).CreateEventResult(context.Background(), entities.CreateEventResultParams{
		EventID:        eventID,
		UserID:         userID,
		BestLapTime:    45000,
		AverageLapTime: 47000,
		Position:       position,
		NumberOfLaps:   12,
	})
	if err != nil {
		t.Fatalf("create result: %v", err)
	}
}
