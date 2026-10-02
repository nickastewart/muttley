package dashboard_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"muttley/dashboard"
	"muttley/event"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/headtohead"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestDashboardShowsTotalHeadToHeadWins(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	bea := createUser(t, db, "Bea", "bea@example.com", "bea")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")
	stranger := createUser(t, db, "Sam", "sam@example.com", "sam")

	addFriend(t, db, ada.ID, grace.ID, "ACCEPTED")
	addFriend(t, db, ada.ID, bea.ID, "ACCEPTED")
	addFriend(t, db, ada.ID, alan.ID, "REQUESTED")

	// Ada is ahead of Grace once. Alan and Sam share that race but are not accepted friends.
	seedSharedRace(t, db, "Whilton Mill", "2024-04-01", map[int64]int64{ada.ID: 1, grace.ID: 4, alan.ID: 2, stranger.ID: 5})
	seedSharedRace(t, db, "Whilton Mill", "2024-04-02", map[int64]int64{ada.ID: 5, grace.ID: 2})
	seedSharedRace(t, db, "Whilton Mill", "2024-04-03", map[int64]int64{ada.ID: 3, grace.ID: 3})
	seedSharedRace(t, db, "Whilton Mill", "2024-05-01", map[int64]int64{ada.ID: 2, bea.ID: 4})
	seedSharedRace(t, db, "Whilton Mill", "2024-05-02", map[int64]int64{ada.ID: 1, bea.ID: 6})

	body := requestDashboard(t, db, &ada).Body.String()
	if !strings.Contains(body, `Head To Head Wins</span></div><div class="dashboard-top-row-item-content"><span>3</span>`) {
		t.Fatalf("dashboard head to head wins = %s, want 3", body)
	}
}

func TestDashboardHeadToHeadWinsAreZeroWithoutSharedFriendRaces(t *testing.T) {
	db := testdb.Open(t)
	otto := createUser(t, db, "Otto", "otto@example.com", "otto")

	body := requestDashboard(t, db, &otto).Body.String()
	if !strings.Contains(body, `Head To Head Wins</span></div><div class="dashboard-top-row-item-content"><span>0</span>`) {
		t.Fatalf("dashboard head to head wins = %s, want 0", body)
	}
}

func TestDashboardBestTrackIsBlankWithoutThreeRaces(t *testing.T) {
	db := testdb.Open(t)
	otto := createUser(t, db, "Otto", "otto-best@example.com", "otto-best")
	seedRace(t, db, otto.ID, "Short Track", "2024-07-01", 1)
	seedRace(t, db, otto.ID, "Short Track", "2024-07-02", 2)

	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	previous := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previous) })

	body := requestDashboard(t, db, &otto).Body.String()
	if !strings.Contains(body, `dashboard-top-row-item-content-best-track">-</div>`) {
		t.Fatalf("best track = %s, want -", body)
	}
	if strings.Contains(logs.String(), "no rows in result set") || strings.Contains(logs.String(), "Error getting bestTrack") {
		t.Fatalf("missing best track was logged as an error: %s", logs.String())
	}
}

func TestDashboardShowsBestTrackAfterThreeRaces(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada-best@example.com", "ada-best")
	seedRace(t, db, ada.ID, "Fast Track", "2024-08-01", 1)
	seedRace(t, db, ada.ID, "Fast Track", "2024-08-02", 2)
	seedRace(t, db, ada.ID, "Fast Track", "2024-08-03", 1)

	body := requestDashboard(t, db, &ada).Body.String()
	if !strings.Contains(body, `dashboard-top-row-item-content-best-track">Fast Track</div>`) {
		t.Fatalf("best track = %s, want Fast Track", body)
	}
}

func requestDashboard(t *testing.T, db *sql.DB, user *entities.User) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	queries := entities.New(db)
	handler := dashboard.NewDashboardHander(
		dashboard.NewDashboardRepository(queries),
		event.NewEventRepository(queries),
		headtohead.NewHeadToHeadRepository(queries),
	)

	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	router.GET("/dashboard", func(c *gin.Context) {
		if user != nil {
			c.Set("currentUser", *user)
		}
		handler.GetDashboard(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	return recorder
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

func seedSharedRace(t *testing.T, db *sql.DB, track, date string, positions map[int64]int64) {
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

type testTemplRender struct {
	Code int
	Data templ.Component
}

func (t testTemplRender) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	w.WriteHeader(t.Code)
	if t.Data != nil {
		return t.Data.Render(context.Background(), w)
	}
	return nil
}

func (t testTemplRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (t *testTemplRender) Instance(name string, data interface{}) render.Render {
	if component, ok := data.(templ.Component); ok {
		return &testTemplRender{Code: http.StatusOK, Data: component}
	}
	return nil
}
