package records_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"muttley/records"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestRecordsPageShowsLapTimes(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	addFriend(t, db, grace, ada, "ACCEPTED")
	seedResult(t, db, "Whilton Mill", "2024-06-01", ada, 48000, 51000)
	seedResult(t, db, "Whilton Mill", "2024-04-01", ada, 49000, 49500)
	seedResult(t, db, "Whilton Mill", "2024-02-01", grace, 47000, 49000)

	recorder := requestRecords(t, db, &entities.User{ID: ada})
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || recorder.Header().Get("HX-Redirect") != "/records" {
		t.Fatalf("status = %d redirect = %q body = %s", recorder.Code, recorder.Header().Get("HX-Redirect"), body)
	}
	if !strings.Contains(body, `href="/styles/records.css"`) || !strings.Contains(body, `hx-get="/records"`) {
		t.Fatalf("page is missing its stylesheet or nav link: %s", body)
	}
	if !strings.Contains(body, "Whilton Mill") || !strings.Contains(body, "00:48.000") || !strings.Contains(body, "2024-06-01") {
		t.Fatalf("page is missing the personal best: %s", body)
	}
	if !strings.Contains(body, "00:49.500") || !strings.Contains(body, "2024-04-01") || !strings.Contains(body, "00:03.000") {
		t.Fatalf("page is missing the average or consistency: %s", body)
	}
	if !strings.Contains(body, "00:47.000") || !strings.Contains(body, "Grace") || !strings.Contains(body, "Grace Racer") {
		t.Fatalf("page is missing the track record: %s", body)
	}
	if strings.Contains(body, "No results yet.") {
		t.Fatalf("page showed the empty state: %s", body)
	}
}

func TestRecordsPageEmptyState(t *testing.T) {
	db := testdb.Open(t)
	otto := createUser(t, db, "Otto", "otto@example.com", "otto")

	body := requestRecords(t, db, &entities.User{ID: otto}).Body.String()
	if !strings.Contains(body, "No results yet.") {
		t.Fatalf("empty page = %s", body)
	}
}

func TestRecordsRequiresAUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/records", records.NewRecordsHandler(failingRepository{}).Records)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/records", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "User is not authenticated") {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestRecordsReportsAQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/records", func(c *gin.Context) {
		c.Set("currentUser", entities.User{ID: 1})
		records.NewRecordsHandler(failingRepository{}).Records(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/records", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "database closed") {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func requestRecords(t *testing.T, db *sql.DB, user *entities.User) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	handler := records.NewRecordsHandler(records.NewRecordsRepository(entities.New(db)))
	router.GET("/records", func(c *gin.Context) {
		if user != nil {
			c.Set("currentUser", *user)
		}
		handler.Records(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/records", nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

type failingRepository struct{}

func (failingRepository) ListLocationRecords(context.Context, int64) ([]entities.ListLocationRecordsRow, error) {
	return nil, errors.New("database closed")
}

func (failingRepository) GetLocationRecordSnapshot(context.Context, entities.GetLocationRecordSnapshotParams) (entities.GetLocationRecordSnapshotRow, error) {
	return entities.GetLocationRecordSnapshotRow{}, errors.New("database closed")
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
