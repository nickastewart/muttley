package fileupload

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/records"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"

	"github.com/nickastewart/muttley-parser/model"
)

func TestSaveEventRejectsPositionOutsideDriverTimes(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	currentUser := createUploadUser(t, db)
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	for _, position := range []int{0, 2} {
		_, _, _, _, err := handler.saveEvent(ctx, currentUser, parsedSession("Whilton Mill", "2024-06-01", "Rental", 1, position))
		if err == nil {
			t.Fatalf("position %d: expected error", position)
		}
		if !errors.Is(err, errBadResultsFile) {
			t.Fatalf("position %d error = %v, want bad results file", position, err)
		}
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

func TestSaveEventCommitsLocationEventAndResult(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	currentUser := createUploadUser(t, db)
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	_, _, saved, _, err := handler.saveEvent(ctx, currentUser, parsedEvent("Whilton Mill"))
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

	_, _, again, _, err := handler.saveEvent(ctx, currentUser, parsedEvent("Whilton Mill"))
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

func TestSaveEventAttachesFriendToSameSession(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	ada := createUploadUser(t, db)
	grace := createUploadUserNamed(t, db, "Grace", "grace@example.com", "grace")
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	adaFile := parsedSession("Daytona Milton Keynes", "16 Jul 2024", "(DMAX Sprint Race)", 8, 1)
	_, adaEvent, adaResult, _, err := handler.saveEvent(ctx, ada, adaFile)
	if err != nil {
		t.Fatalf("save ada: %v", err)
	}

	graceFile := parsedSession("Daytona Milton Keynes", "16 Jul 2024", "(DMAX Sprint Race)", 8, 4)
	_, graceEvent, graceResult, _, err := handler.saveEvent(ctx, grace, graceFile)
	if err != nil {
		t.Fatalf("save grace: %v", err)
	}
	if graceEvent.ID != adaEvent.ID {
		t.Fatalf("grace event id = %d, want ada event %d", graceEvent.ID, adaEvent.ID)
	}
	if graceResult.EventID != adaResult.EventID || graceResult.UserID != grace.ID || graceResult.Position != 4 {
		t.Fatalf("grace result = %+v", graceResult)
	}
	if got := countRows(t, db, "event"); got != 1 {
		t.Fatalf("events = %d, want 1", got)
	}
	if got := countRows(t, db, "event_result"); got != 2 {
		t.Fatalf("results = %d, want 2", got)
	}
}

func TestSaveEventUsesTotalDriversToSeparateSameDayHeats(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	ada := createUploadUser(t, db)
	grace := createUploadUserNamed(t, db, "Grace", "grace@example.com", "grace")
	bea := createUploadUserNamed(t, db, "Bea", "bea@example.com", "bea")
	handler := newUploadHandler(db, eventresult.NewEventResultRepository(entities.New(db)))

	_, sprint, _, _, err := handler.saveEvent(ctx, ada, parsedSession("Daytona Milton Keynes", "16 Jul 2024", "(DMAX Sprint Race)", 8, 1))
	if err != nil {
		t.Fatalf("save sprint: %v", err)
	}
	_, endurance, _, _, err := handler.saveEvent(ctx, grace, parsedSession("Daytona Milton Keynes", "16 Jul 2024", "(DMAX Sprint Race)", 12, 2))
	if err != nil {
		t.Fatalf("save endurance: %v", err)
	}
	if endurance.ID == sprint.ID {
		t.Fatal("different driver counts created one event")
	}
	if got := countRows(t, db, "event"); got != 2 {
		t.Fatalf("events = %d, want 2", got)
	}

	_, joined, result, _, err := handler.saveEvent(ctx, bea, parsedSession("Daytona Milton Keynes", "16 Jul 2024", "(DMAX Sprint Race)", 12, 6))
	if err != nil {
		t.Fatalf("save bea: %v", err)
	}
	if joined.ID != endurance.ID {
		t.Fatalf("bea event id = %d, want endurance %d", joined.ID, endurance.ID)
	}
	if result.Position != 6 || result.EventID != endurance.ID {
		t.Fatalf("bea result = %+v", result)
	}
	if got := countRows(t, db, "event"); got != 2 {
		t.Fatalf("events after third upload = %d, want 2", got)
	}
	if got := countRows(t, db, "event_result"); got != 3 {
		t.Fatalf("results = %d, want 3", got)
	}
}

func TestMatchSession(t *testing.T) {
	sprint := entities.Event{ID: 1, TotalDrivers: 8}
	endurance := entities.Event{ID: 2, TotalDrivers: 12}
	candidates := []entities.Event{sprint, endurance}

	got, ok := matchSession(candidates, 12)
	if !ok || got.ID != endurance.ID {
		t.Fatalf("match 12 = %+v ok=%v, want endurance", got, ok)
	}
	got, ok = matchSession(candidates, 8)
	if !ok || got.ID != sprint.ID {
		t.Fatalf("match 8 = %+v ok=%v, want sprint", got, ok)
	}
	if _, ok := matchSession(candidates, 10); ok {
		t.Fatal("unmatched driver count should not attach")
	}
	got, ok = matchSession(candidates, 0)
	if !ok || got.ID != sprint.ID {
		t.Fatalf("missing driver count = %+v ok=%v, want oldest", got, ok)
	}
	if _, ok := matchSession(nil, 8); ok {
		t.Fatal("no candidates should not attach")
	}
}

func TestSaveEventRollsBackNewLocationAndEvent(t *testing.T) {
	db := testdb.Open(t)
	handler := newUploadHandler(db, failingResults{})

	_, _, _, _, err := handler.saveEvent(context.Background(), entities.User{ID: 1}, parsedEvent("New Track"))
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

	_, _, _, _, err := handler.saveEvent(ctx, entities.User{ID: 1}, parsedEvent("Whilton Mill"))
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
	queries := entities.New(db)
	return NewFileUploadHandler(
		user.NewUserRepository(db),
		event.NewEventRepository(queries),
		location.NewLocationRepository(queries),
		results,
		records.NewRecordsRepository(queries),
		sqlite.NewTransactor(db),
	)
}

func parsedEvent(locationName string) *model.Event {
	return parsedSession(locationName, "2024-06-01", "Rental", 1, 1)
}

func parsedSession(locationName, date, raceType string, drivers, position int) *model.Event {
	times := make([]model.DriverTime, drivers)
	for i := range times {
		times[i] = model.DriverTime{
			Pos:    i + 1,
			Kart:   "1",
			Racer:  "Driver",
			Best:   45000,
			NoLaps: 10,
			Avg:    47000,
		}
	}
	return &model.Event{
		Date:     date,
		Location: locationName,
		RaceType: raceType,
		DriverInfo: model.DriverInfo{
			Name:     "Ada",
			Position: position,
		},
		DriverTimes: times,
	}
}

func createUploadUser(t *testing.T, db *sql.DB) entities.User {
	t.Helper()
	return createUploadUserNamed(t, db, "Ada", "ada@example.com", "ada")
}

func createUploadUserNamed(t *testing.T, db *sql.DB, firstName, email, profileID string) entities.User {
	t.Helper()
	repo := user.NewUserRepository(db)
	ctx := context.Background()
	if _, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    "Lovelace",
		Email:       email,
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
