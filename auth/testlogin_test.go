package auth_test

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"muttley/auth"

	"github.com/gin-gonic/gin"
)

func TestTestLoginSetsTheAccessTokenCookie(t *testing.T) {
	t.Setenv("PROFILE", "test")
	handler, db := newAuthHandler(t, &captureMailer{})
	router := testRouter(handler)
	auth.RegisterTestLogin(router, handler)
	createAuthUser(t, db, "ada@example.com")
	id := authUserID(t, db, "ada@example.com")

	invalid := request(t, router, http.MethodGet, "/test/login/nope", nil, nil)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid id status = %d body = %s", invalid.Code, invalid.Body.String())
	}
	unknown := request(t, router, http.MethodGet, "/test/login/9999", nil, nil)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown user status = %d body = %s", unknown.Code, unknown.Body.String())
	}

	rec := request(t, router, http.MethodGet, "/test/login/"+strconv.FormatInt(id, 10), nil, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), strconv.FormatInt(id, 10)) {
		t.Fatalf("login status = %d body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].Value == "" || !cookies[0].HttpOnly {
		t.Fatalf("cookies = %+v", cookies)
	}
	home := request(t, router, http.MethodGet, "/", nil, cookies)
	if home.Code != http.StatusOK || home.Body.String() != "ok" {
		t.Fatalf("home status = %d body = %s", home.Code, home.Body.String())
	}

	t.Setenv("PROFILE", "")
	blocked := request(t, router, http.MethodGet, "/test/login/"+strconv.FormatInt(id, 10), nil, nil)
	if blocked.Code != http.StatusNotFound || len(blocked.Result().Cookies()) != 0 {
		t.Fatalf("blocked status = %d cookies = %+v body = %s", blocked.Code, blocked.Result().Cookies(), blocked.Body.String())
	}
}

func TestTestLoginRouteHiddenUnlessTestProfile(t *testing.T) {
	t.Setenv("PROFILE", "")
	handler, db := newAuthHandler(t, &captureMailer{})
	router := testRouter(handler)
	auth.RegisterTestLogin(router, handler)
	createAuthUser(t, db, "ada@example.com")
	id := authUserID(t, db, "ada@example.com")

	hidden := request(t, router, http.MethodGet, "/test/login/"+strconv.FormatInt(id, 10), nil, nil)
	if hidden.Code != http.StatusNotFound || len(hidden.Result().Cookies()) != 0 {
		t.Fatalf("hidden status = %d cookies = %+v body = %s", hidden.Code, hidden.Result().Cookies(), hidden.Body.String())
	}

	t.Setenv("PROFILE", "test")
	visible := gin.New()
	auth.RegisterTestLogin(visible, handler)
	visible.GET("/", handler.CheckAccessToken, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	found := request(t, visible, http.MethodGet, "/test/login/"+strconv.FormatInt(id, 10), nil, nil)
	if found.Code != http.StatusOK {
		t.Fatalf("visible status = %d body = %s", found.Code, found.Body.String())
	}
	cookies := found.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].Value == "" {
		t.Fatalf("cookies = %+v", cookies)
	}
	home := request(t, visible, http.MethodGet, "/", nil, cookies)
	if home.Code != http.StatusOK || home.Body.String() != "ok" {
		t.Fatalf("home status = %d body = %s", home.Code, home.Body.String())
	}
}

func authUserID(t *testing.T, db *sql.DB, email string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(`SELECT id FROM user WHERE email = ?`, email).Scan(&id); err != nil {
		t.Fatalf("user id: %v", err)
	}
	if id == 0 {
		t.Fatal("user id = 0")
	}
	return id
}
