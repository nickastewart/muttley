package headtohead_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"muttley/headtohead"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestHeadToHeadPageShowsTotals(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	addFriend(t, db, ada, grace, "ACCEPTED")
	seedEvent(t, db, "Whilton Mill", "2024-04-01", map[int64]int64{ada: 1, grace: 4})
	seedEvent(t, db, "Whilton Mill", "2024-04-02", map[int64]int64{ada: 3, grace: 3})

	recorder := requestHeadToHead(t, db, &entities.User{ID: ada})
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || recorder.Header().Get("HX-Redirect") != "/head-to-head" {
		t.Fatalf("status = %d redirect = %q body = %s", recorder.Code, recorder.Header().Get("HX-Redirect"), body)
	}
	if !strings.Contains(body, `href="/styles/headtohead.css"`) || !strings.Contains(body, `hx-get="/head-to-head"`) {
		t.Fatalf("page is missing its stylesheet or nav link: %s", body)
	}
	if !strings.Contains(body, "Grace") || !strings.Contains(body, "Grace Racer") {
		t.Fatalf("page is missing the friend: %s", body)
	}
	if !strings.Contains(body, ">1<") || !strings.Contains(body, ">0<") || !strings.Contains(body, ">2<") {
		t.Fatalf("page totals = %s, want 1 win, 0 losses, 2 races", body)
	}
	if strings.Contains(body, "No shared races with friends yet.") {
		t.Fatalf("page showed the empty state: %s", body)
	}
}

func TestHeadToHeadPageEmptyState(t *testing.T) {
	db := testdb.Open(t)
	otto := createUser(t, db, "Otto", "otto@example.com", "otto")

	body := requestHeadToHead(t, db, &entities.User{ID: otto}).Body.String()
	if !strings.Contains(body, "No shared races with friends yet.") {
		t.Fatalf("empty page = %s", body)
	}
}

func TestHeadToHeadRequiresAUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/head-to-head", headtohead.NewHeadToHeadHandler(failingRepository{}).HeadToHead)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/head-to-head", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "User is not authenticated") {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHeadToHeadReportsAQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/head-to-head", func(c *gin.Context) {
		c.Set("currentUser", entities.User{ID: 1})
		headtohead.NewHeadToHeadHandler(failingRepository{}).HeadToHead(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/head-to-head", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "database closed") {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func requestHeadToHead(t *testing.T, db *sql.DB, user *entities.User) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	handler := headtohead.NewHeadToHeadHandler(headtohead.NewHeadToHeadRepository(entities.New(db)))
	router.GET("/head-to-head", func(c *gin.Context) {
		if user != nil {
			c.Set("currentUser", *user)
		}
		handler.HeadToHead(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/head-to-head", nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

type failingRepository struct{}

func (failingRepository) GetHeadToHead(context.Context, int64) ([]entities.GetHeadToHeadRow, error) {
	return nil, errors.New("database closed")
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
