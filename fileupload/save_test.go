package fileupload

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"

	"github.com/nickastewart/muttley-parser/model"
)

func TestSaveEventCommitsLocationEventAndResult(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	currentUser := createUploadUser(t, db)
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	saved, err := handler.saveEvent(ctx, currentUser, parsedEvent("Whilton Mill"))
	if err != nil {
		t.Fatalf("save event: %v", err)
	}
	if saved.UserID != currentUser.ID || saved.Position != 1 {
		t.Fatalf("saved result = %+v", saved)
	}
	if got := countRows(t, db, "location"); got != 1 {
		t.Fatalf("locations = %d, want 1", got)
	}
	if got := countRows(t, db, "event"); got != 1 {
		t.Fatalf("events = %d, want 1", got)
	}
	if got := countRows(t, db, "event_result"); got != 1 {
		t.Fatalf("results = %d, want 1", got)
	}

	again, err := handler.saveEvent(ctx, currentUser, parsedEvent("Whilton Mill"))
	if err != nil {
		t.Fatalf("save event again: %v", err)
	}
	if again.ID != saved.ID {
		t.Fatalf("second save id = %d, want existing %d", again.ID, saved.ID)
	}
	if got := countRows(t, db, "location"); got != 1 {
		t.Fatalf("locations after second save = %d, want 1", got)
	}
	if got := countRows(t, db, "event"); got != 1 {
		t.Fatalf("events after second save = %d, want 1", got)
	}
	if got := countRows(t, db, "event_result"); got != 1 {
		t.Fatalf("results after second save = %d, want 1", got)
	}
}

func TestSaveEventRollsBackNewLocationAndEvent(t *testing.T) {
	db := testdb.Open(t)
	handler := newUploadHandler(db, failingResults{})

	_, err := handler.saveEvent(context.Background(), entities.GetUserByIdRow{ID: 1}, parsedEvent("New Track"))
	if err == nil {
		t.Fatal("expected save to fail")
	}
	if got := countRows(t, db, "location"); got != 0 {
		t.Fatalf("locations = %d, want 0", got)
	}
	if got := countRows(t, db, "event"); got != 0 {
		t.Fatalf("events = %d, want 0", got)
	}
	if got := countRows(t, db, "event_result"); got != 0 {
		t.Fatalf("results = %d, want 0", got)
	}
}

func TestSaveEventKeepsExistingLocationWhenResultFails(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	if _, err := location.NewLocationRepository(entities.New(db)).CreateLocation(ctx, "Whilton Mill"); err != nil {
		t.Fatalf("seed location: %v", err)
	}
	handler := newUploadHandler(db, failingResults{})

	_, err := handler.saveEvent(ctx, entities.GetUserByIdRow{ID: 1}, parsedEvent("Whilton Mill"))
	if err == nil {
		t.Fatal("expected save to fail")
	}
	if got := countRows(t, db, "location"); got != 1 {
		t.Fatalf("locations = %d, want 1", got)
	}
	if got := countRows(t, db, "event"); got != 0 {
		t.Fatalf("events = %d, want 0", got)
	}
	if got := countRows(t, db, "event_result"); got != 0 {
		t.Fatalf("results = %d, want 0", got)
	}
}

func newUploadHandler(db *sql.DB, results eventresult.EventResultRepository) *FileUploadHandler {
	return NewFileUploadHandler(
		user.NewUserRepository(db),
		event.NewEventRepository(entities.New(db)),
		location.NewLocationRepository(entities.New(db)),
		results,
		sqlite.NewTransactor(db),
	)
}

func parsedEvent(locationName string) *model.Event {
	return &model.Event{
		Date:     "2024-06-01",
		Location: locationName,
		RaceType: "Rental",
		DriverInfo: model.DriverInfo{
			Name:     "Ada",
			Position: 1,
		},
		DriverTimes: []model.DriverTime{{
			Pos:    1,
			Kart:   "1",
			Racer:  "Ada",
			Best:   45000,
			NoLaps: 10,
			Avg:    47000,
		}},
	}
}

func createUploadUser(t *testing.T, db *sql.DB) entities.GetUserByIdRow {
	t.Helper()
	repo := user.NewUserRepository(db)
	ctx := context.Background()
	if _, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
		Password:    "secret",
		ProfileID:   "ada",
		DisplayName: "Ada",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	byEmail, err := repo.GetUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	byID, err := repo.GetUserById(ctx, byEmail.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	return byID
}

type failingResults struct{}

func (failingResults) CreateEventResult(context.Context, entities.CreateEventResultParams) (entities.EventResult, error) {
	return entities.EventResult{}, errors.New("result failed")
}

func (failingResults) GetEventResultByEventIdAndUserId(context.Context, entities.GetEventResultByEventIdAndUserIdParams) (entities.EventResult, error) {
	return entities.EventResult{}, sql.ErrNoRows
}

func (failingResults) GetUserFriendsResults(context.Context, int64) ([]entities.GetUserFriendsResultsRow, error) {
	return nil, nil
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var query string
	switch table {
	case "location":
		query = "SELECT COUNT(*) FROM location"
	case "event":
		query = "SELECT COUNT(*) FROM event"
	case "event_result":
		query = "SELECT COUNT(*) FROM event_result"
	default:
		t.Fatalf("unknown table %s", table)
	}
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
