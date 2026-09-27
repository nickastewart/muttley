package headtohead_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/headtohead"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestGetHeadToHeadCountsSharedRaces(t *testing.T) {
	db := testdb.Open(t)
	repo := headtohead.NewHeadToHeadRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	bea := createUser(t, db, "Bea", "bea@example.com", "bea")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")
	nia := createUser(t, db, "Nia", "nia@example.com", "nia")
	stranger := createUser(t, db, "Sam", "sam@example.com", "sam")
	otto := createUser(t, db, "Otto", "otto@example.com", "otto")

	// Stored from Grace's side so the lookup has to follow both friendship directions.
	addFriend(t, db, grace, ada, "ACCEPTED")
	addFriend(t, db, ada, grace, "ACCEPTED")
	addFriend(t, db, ada, bea, "ACCEPTED")
	addFriend(t, db, ada, alan, "REQUESTED")
	addFriend(t, db, ada, nia, "CANCELLED")

	seedEvent(t, db, "Whilton Mill", "2024-04-01", map[int64]int64{ada: 1, grace: 4, alan: 2, nia: 3, stranger: 5})
	seedEvent(t, db, "Whilton Mill", "2024-04-02", map[int64]int64{ada: 5, grace: 2})
	seedEvent(t, db, "Whilton Mill", "2024-04-03", map[int64]int64{ada: 3, grace: 3})
	seedEvent(t, db, "Whilton Mill", "2024-04-04", map[int64]int64{ada: 1})
	seedEvent(t, db, "Whilton Mill", "2024-04-05", map[int64]int64{grace: 1})
	seedEvent(t, db, "Whilton Mill", "2024-05-01", map[int64]int64{ada: 2, bea: 4})
	seedEvent(t, db, "Whilton Mill", "2024-05-02", map[int64]int64{ada: 1, bea: 6})

	ctx := context.Background()
	rows, err := repo.GetHeadToHead(ctx, ada)
	if err != nil {
		t.Fatalf("get head to head: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want Grace and Bea", rows)
	}
	if rows[0].ID != grace || rows[1].ID != bea {
		t.Fatalf("order = %+v, want the friend with more shared races first", rows)
	}

	graceRow := rows[0]
	if graceRow.DisplayName != "Grace" || graceRow.FirstName != "Grace" || graceRow.LastName != "Racer" {
		t.Fatalf("grace = %+v", graceRow)
	}
	// Ada finished ahead once, behind once, and tied once. The tie is a shared race for neither driver.
	if graceRow.UserWins != 1 || graceRow.FriendWins != 1 || graceRow.SharedRaces != 3 {
		t.Fatalf("ada vs grace = %+v", graceRow)
	}
	if rows[1].UserWins != 2 || rows[1].FriendWins != 0 || rows[1].SharedRaces != 2 {
		t.Fatalf("ada vs bea = %+v", rows[1])
	}

	fromGrace, err := repo.GetHeadToHead(ctx, grace)
	if err != nil {
		t.Fatalf("get grace head to head: %v", err)
	}
	if len(fromGrace) != 1 || fromGrace[0].ID != ada || fromGrace[0].UserWins != 1 || fromGrace[0].FriendWins != 1 || fromGrace[0].SharedRaces != 3 {
		t.Fatalf("grace vs ada = %+v", fromGrace)
	}

	fromBea, err := repo.GetHeadToHead(ctx, bea)
	if err != nil {
		t.Fatalf("get bea head to head: %v", err)
	}
	if len(fromBea) != 1 || fromBea[0].ID != ada || fromBea[0].UserWins != 0 || fromBea[0].FriendWins != 2 || fromBea[0].SharedRaces != 2 {
		t.Fatalf("bea vs ada = %+v", fromBea)
	}

	none, err := repo.GetHeadToHead(ctx, otto)
	if err != nil {
		t.Fatalf("get empty head to head: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("empty head to head = %+v", none)
	}
}

func createUser(t *testing.T, db *sql.DB, firstName, email, profileID string) int64 {
	t.Helper()
	repo := user.NewUserRepository(db)
	ctx := context.Background()
	if _, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    "Racer",
		Email:       email,
		ProfileID:   profileID,
		DisplayName: firstName,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	created, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	return created.ID
}

func addFriend(t *testing.T, db *sql.DB, userID, friendID int64, status string) {
	t.Helper()
	if _, err := friend.NewFriendRepository(entities.New(db)).AddFriend(context.Background(), entities.AddFriendParams{
		UserID:       userID,
		FriendID:     friendID,
		FriendStatus: status,
	}); err != nil {
		t.Fatalf("add friend: %v", err)
	}
}

func seedEvent(t *testing.T, db *sql.DB, track, date string, positions map[int64]int64) {
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
		TotalDrivers: int64(len(positions)),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	results := eventresult.NewEventResultRepository(entities.New(db))
	for userID, position := range positions {
		if _, err := results.CreateEventResult(ctx, entities.CreateEventResultParams{
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
}
