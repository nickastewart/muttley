package fileupload

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"muttley/eventresult"
	"muttley/friend"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/templates"

	"github.com/nickastewart/muttley-parser/model"
)

func TestRecordNotes(t *testing.T) {
	if got := recordNotes(false, 44000, 0, 0); got.PersonalBest || got.TrackRecord {
		t.Fatalf("existing result = %+v, want no record", got)
	}
	if got := recordNotes(true, 0, 0, 0); got.PersonalBest || got.TrackRecord {
		t.Fatalf("zero lap = %+v, want no record", got)
	}
	if got := recordNotes(true, 45000, 0, 0); !got.PersonalBest || !got.TrackRecord {
		t.Fatalf("first lap = %+v, want both records", got)
	}
	if got := recordNotes(true, 44000, 45000, 43000); !got.PersonalBest || got.TrackRecord {
		t.Fatalf("personal best only = %+v", got)
	}
	if got := recordNotes(true, 42000, 45000, 43000); !got.PersonalBest || !got.TrackRecord {
		t.Fatalf("both beaten = %+v", got)
	}
	if got := recordNotes(true, 45000, 45000, 45000); got.PersonalBest || got.TrackRecord {
		t.Fatalf("equal time = %+v, want no record", got)
	}
}

func TestSaveEventFlagsPersonalBestAndTrackRecord(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	ada := createUploadUser(t, db)
	grace := createUploadUserNamed(t, db, "Grace", "grace@example.com", "grace")
	alan := createUploadUserNamed(t, db, "Alan", "alan@example.com", "alan")
	addUploadFriend(t, db, grace.ID, ada.ID, "ACCEPTED")
	addUploadFriend(t, db, ada.ID, alan.ID, "REQUESTED")
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	_, _, _, first, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-01", "Rental", 1, 1, 45000, 47000))
	if err != nil {
		t.Fatalf("first save: %v", err)
	}
	if !first.PersonalBest || !first.TrackRecord {
		t.Fatalf("first save notes = %+v, want both records", first)
	}

	_, _, _, again, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-01", "Rental", 1, 1, 45000, 47000))
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if again.PersonalBest || again.TrackRecord {
		t.Fatalf("repeat save notes = %+v, want no record", again)
	}

	_, _, _, slower, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-02", "Rental", 1, 1, 46000, 48000))
	if err != nil {
		t.Fatalf("slower save: %v", err)
	}
	if slower.PersonalBest || slower.TrackRecord {
		t.Fatalf("slower notes = %+v, want no record", slower)
	}

	if _, _, _, _, err := handler.saveEvent(ctx, grace, parsedLap("Whilton Mill", "2024-06-03", "Rental", 1, 1, 43000, 45000)); err != nil {
		t.Fatalf("grace save: %v", err)
	}
	if _, _, _, _, err := handler.saveEvent(ctx, alan, parsedLap("Whilton Mill", "2024-06-04", "Rental", 1, 1, 42000, 44000)); err != nil {
		t.Fatalf("alan save: %v", err)
	}

	_, _, _, personal, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-05", "Rental", 1, 1, 44000, 46000))
	if err != nil {
		t.Fatalf("personal save: %v", err)
	}
	if !personal.PersonalBest || personal.TrackRecord {
		t.Fatalf("personal notes = %+v, want personal best only", personal)
	}

	_, _, _, both, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-06", "Rental", 1, 1, 41000, 43000))
	if err != nil {
		t.Fatalf("record save: %v", err)
	}
	if !both.PersonalBest || !both.TrackRecord {
		t.Fatalf("record notes = %+v, want both records", both)
	}

	_, _, _, missing, err := handler.saveEvent(ctx, ada, parsedLap("Whilton Mill", "2024-06-07", "Rental", 1, 1, 0, 47000))
	if err != nil {
		t.Fatalf("zero save: %v", err)
	}
	if missing.PersonalBest || missing.TrackRecord {
		t.Fatalf("zero notes = %+v, want no record", missing)
	}
}

func TestUploadSuccessShowsRecords(t *testing.T) {
	var buf bytes.Buffer
	component := templates.UploadSuccess(
		entities.Location{Name: "Whilton Mill"},
		entities.Event{Date: "2024-06-01", Type: "Rental"},
		entities.EventResult{BestLapTime: 41000, AverageLapTime: 43000, Position: 1, NumberOfLaps: 10},
		true,
		true,
	)
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	body := buf.String()
	if !containsAll(body, "Personal best", "Track record", "00:41.000") {
		t.Fatalf("success page = %s", body)
	}

	buf.Reset()
	plain := templates.UploadSuccess(
		entities.Location{Name: "Whilton Mill"},
		entities.Event{Date: "2024-06-01", Type: "Rental"},
		entities.EventResult{BestLapTime: 45000, AverageLapTime: 47000, Position: 1, NumberOfLaps: 10},
		false,
		false,
	)
	if err := plain.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render plain: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte("Personal best")) || bytes.Contains(buf.Bytes(), []byte("Track record")) {
		t.Fatalf("plain success page showed a record: %s", buf.String())
	}
}

func parsedLap(locationName, date, raceType string, drivers, position, best, avg int) *model.Event {
	event := parsedSession(locationName, date, raceType, drivers, position)
	event.DriverTimes[position-1].Best = best
	event.DriverTimes[position-1].Avg = avg
	return event
}

func addUploadFriend(t *testing.T, db *sql.DB, userID, friendID int64, status string) {
	t.Helper()
	if _, err := friend.NewFriendRepository(entities.New(db)).AddFriend(context.Background(), entities.AddFriendParams{
		UserID:       userID,
		FriendID:     friendID,
		FriendStatus: status,
	}); err != nil {
		t.Fatalf("add friend: %v", err)
	}
}

func containsAll(body string, parts ...string) bool {
	for _, part := range parts {
		if !bytes.Contains([]byte(body), []byte(part)) {
			return false
		}
	}
	return true
}
