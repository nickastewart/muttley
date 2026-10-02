package records_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/location"
	"muttley/records"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestListLocationRecords(t *testing.T) {
	db := testdb.Open(t)
	repo := records.NewRecordsRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")
	stranger := createUser(t, db, "Sam", "sam@example.com", "sam")

	addFriend(t, db, grace, ada, "ACCEPTED")
	addFriend(t, db, ada, alan, "REQUESTED")

	seedResult(t, db, "Old Track", "2024-01-01", ada, 70000, 71000)
	seedResult(t, db, "Solo", "2024-04-01", ada, 60000, 63000)
	seedResult(t, db, "Solo", "2024-05-01", ada, 60000, 62000)
	seedResult(t, db, "New Track", "2024-03-01", ada, 50000, 52000)
	seedResult(t, db, "New Track", "2024-04-01", ada, 49000, 49500)
	seedResult(t, db, "New Track", "2024-06-01", ada, 48000, 51000)
	seedResult(t, db, "New Track", "2024-02-01", grace, 47000, 49000)
	seedResult(t, db, "New Track", "2024-02-02", alan, 10000, 11000)
	seedResult(t, db, "New Track", "2024-02-03", stranger, 9000, 10000)
	seedResult(t, db, "Friend Only", "2024-08-01", grace, 30000, 31000)
	seedResult(t, db, "Wet", "2024-07-01", ada, 0, 0)

	rows, err := repo.ListLocationRecords(context.Background(), ada)
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows = %+v, want Wet, New Track, Solo, Old Track", rows)
	}
	if rows[0].LocationName != "Wet" || rows[1].LocationName != "New Track" || rows[2].LocationName != "Solo" || rows[3].LocationName != "Old Track" {
		t.Fatalf("order = %+v", rows)
	}

	wet := rows[0]
	if wet.PersonalBestLap != 0 || wet.BestAverageLap != 0 || wet.TrackRecordLap != 0 || wet.PersonalBestDate != "" {
		t.Fatalf("zero laps counted as records: %+v", wet)
	}

	newest := rows[1]
	if newest.PersonalBestLap != 48000 || newest.PersonalBestDate != "2024-06-01" || newest.PersonalBestAverageLap != 51000 {
		t.Fatalf("personal best = %+v", newest)
	}
	if newest.BestAverageLap != 49500 || newest.BestAverageDate != "2024-04-01" {
		t.Fatalf("best average = %+v", newest)
	}
	if newest.TrackRecordLap != 47000 || newest.TrackRecordDisplayName != "Grace" || newest.TrackRecordFirstName != "Grace" || newest.TrackRecordLastName != "Racer" {
		t.Fatalf("track record = %+v", newest)
	}

	solo := rows[2]
	if solo.PersonalBestLap != 60000 || solo.PersonalBestDate != "2024-04-01" || solo.PersonalBestAverageLap != 63000 {
		t.Fatalf("tied personal best = %+v", solo)
	}
	if solo.TrackRecordLap != 60000 || solo.TrackRecordDisplayName != "Ada" {
		t.Fatalf("solo track record = %+v", solo)
	}

	old := rows[3]
	if old.PersonalBestLap != 70000 || old.BestAverageLap != 71000 || old.TrackRecordDisplayName != "Ada" {
		t.Fatalf("old track = %+v", old)
	}

	fromGrace, err := repo.ListLocationRecords(context.Background(), grace)
	if err != nil {
		t.Fatalf("list grace: %v", err)
	}
	if len(fromGrace) != 2 || fromGrace[0].LocationName != "Friend Only" || fromGrace[1].LocationName != "New Track" {
		t.Fatalf("grace rows = %+v", fromGrace)
	}
	if fromGrace[1].TrackRecordLap != 47000 || fromGrace[1].TrackRecordDisplayName != "Grace" {
		t.Fatalf("grace track record = %+v", fromGrace[1])
	}
}

func TestGetLocationRecordSnapshot(t *testing.T) {
	db := testdb.Open(t)
	repo := records.NewRecordsRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")
	addFriend(t, db, grace, ada, "ACCEPTED")
	addFriend(t, db, ada, alan, "REQUESTED")

	ctx := context.Background()
	empty, err := repo.GetLocationRecordSnapshot(ctx, entities.GetLocationRecordSnapshotParams{
		Userid:     ada,
		Locationid: 1,
	})
	if err != nil {
		t.Fatalf("empty snapshot: %v", err)
	}
	if empty.PersonalBestLap != 0 || empty.TrackRecordLap != 0 {
		t.Fatalf("empty snapshot = %+v", empty)
	}

	locationID := seedResult(t, db, "Whilton Mill", "2024-06-01", ada, 0, 47000)
	seedResult(t, db, "Whilton Mill", "2024-06-02", ada, 45000, 47000)
	seedResult(t, db, "Whilton Mill", "2024-06-03", alan, 40000, 42000)

	beforeFriend, err := repo.GetLocationRecordSnapshot(ctx, entities.GetLocationRecordSnapshotParams{
		Userid:     ada,
		Locationid: locationID,
	})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if beforeFriend.PersonalBestLap != 45000 || beforeFriend.TrackRecordLap != 45000 {
		t.Fatalf("before friend = %+v", beforeFriend)
	}

	seedResult(t, db, "Whilton Mill", "2024-06-04", grace, 43000, 45000)
	afterFriend, err := repo.GetLocationRecordSnapshot(ctx, entities.GetLocationRecordSnapshotParams{
		Userid:     ada,
		Locationid: locationID,
	})
	if err != nil {
		t.Fatalf("snapshot after friend: %v", err)
	}
	if afterFriend.PersonalBestLap != 45000 || afterFriend.TrackRecordLap != 43000 {
		t.Fatalf("after friend = %+v", afterFriend)
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

func seedResult(t *testing.T, db *sql.DB, track, date string, userID, best, avg int64) int64 {
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
		TotalDrivers: 1,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	if _, err := eventresult.NewEventResultRepository(entities.New(db)).CreateEventResult(ctx, entities.CreateEventResultParams{
		EventID:        race.ID,
		UserID:         userID,
		BestLapTime:    best,
		AverageLapTime: avg,
		Position:       1,
		NumberOfLaps:   10,
	}); err != nil {
		t.Fatalf("create result: %v", err)
	}
	return loc.ID
}
