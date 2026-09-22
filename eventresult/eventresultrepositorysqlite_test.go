package eventresult_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestCreateAndGetEventResult(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	results := eventresult.NewEventResultRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	track := createLocation(t, db, "Whilton Mill")
	race := createEvent(t, db, track.ID, "2024-06-01")

	created, err := results.CreateEventResult(ctx, entities.CreateEventResultParams{
		EventID:        race.ID,
		UserID:         ada.ID,
		BestLapTime:    45123,
		AverageLapTime: 47200,
		Position:       2,
		NumberOfLaps:   14,
	})
	if err != nil {
		t.Fatalf("create result: %v", err)
	}
	if created.ID == 0 || created.EventID != race.ID || created.UserID != ada.ID || created.BestLapTime != 45123 || created.AverageLapTime != 47200 || created.Position != 2 || created.NumberOfLaps != 14 {
		t.Fatalf("created result = %+v", created)
	}

	found, err := results.GetEventResultByEventIdAndUserId(ctx, entities.GetEventResultByEventIdAndUserIdParams{
		EventID: race.ID,
		UserID:  ada.ID,
	})
	if err != nil {
		t.Fatalf("get result: %v", err)
	}
	if found.ID != created.ID || found.Position != 2 {
		t.Fatalf("found result = %+v", found)
	}
}

func TestGetMissingEventResult(t *testing.T) {
	results := eventresult.NewEventResultRepository(entities.New(testdb.Open(t)))

	_, err := results.GetEventResultByEventIdAndUserId(context.Background(), entities.GetEventResultByEventIdAndUserIdParams{
		EventID: 1,
		UserID:  1,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func TestGetUserFriendsResults(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	results := eventresult.NewEventResultRepository(entities.New(db))
	friends := friend.NewFriendRepository(entities.New(db))

	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")
	track := createLocation(t, db, "Whilton Mill")
	race := createEvent(t, db, track.ID, "2024-06-01")

	createResult(t, db, race.ID, ada.ID, 1)
	createResult(t, db, race.ID, grace.ID, 4)
	createResult(t, db, race.ID, alan.ID, 6)

	// Results are included only when the caller is friend.user_id. Alan added
	// Ada, so his result is not Ada's friend result.
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       ada.ID,
		FriendID:     grace.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add grace: %v", err)
	}
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       alan.ID,
		FriendID:     ada.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add alan: %v", err)
	}

	rows, err := results.GetUserFriendsResults(ctx, ada.ID)
	if err != nil {
		t.Fatalf("get friend results: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d results, want 1: %+v", len(rows), rows)
	}
	row := rows[0]
	if row.User.ID != grace.ID || row.EventResult.Position != 4 || row.Location.Name != "Whilton Mill" || row.Event.Date != "2024-06-01" {
		t.Fatalf("friend result = %+v", row)
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

func createResult(t *testing.T, db *sql.DB, eventID, userID, position int64) entities.EventResult {
	t.Helper()
	result, err := eventresult.NewEventResultRepository(entities.New(db)).CreateEventResult(context.Background(), entities.CreateEventResultParams{
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
	return result
}
