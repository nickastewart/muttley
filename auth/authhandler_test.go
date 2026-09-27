package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"muttley/auth"
	"muttley/mailer"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestLoginPageHasNoPassword(t *testing.T) {
	body := requestHTML(t, newTestRouter(t, &captureMailer{}), http.MethodGet, "/login", nil, nil)
	if strings.Contains(body, "password") || !strings.Contains(body, "Email me a sign-in link") {
		t.Fatalf("login page = %s", body)
	}
	if !strings.Contains(body, `href="/signup"`) {
		t.Fatalf("login page missing create account link: %s", body)
	}
}

func TestSignupPageIsEmailOnly(t *testing.T) {
	body := requestHTML(t, newTestRouter(t, &captureMailer{}), http.MethodGet, "/signup?email=ada@example.com", nil, nil)
	if strings.Contains(body, "password") || !strings.Contains(body, "Email me a sign-up link") {
		t.Fatalf("signup page = %s", body)
	}
	if !strings.Contains(body, `value="ada@example.com"`) {
		t.Fatalf("signup page did not keep the email: %s", body)
	}
}

func TestLoginRejectsUnknownEmail(t *testing.T) {
	sent := &captureMailer{}
	body := requestHTML(t, newTestRouter(t, sent), http.MethodPost, "/login", url.Values{"email": {"missing@example.com"}}, nil)
	if !strings.Contains(body, "No account found for that email.") {
		t.Fatalf("body = %s", body)
	}
	if !strings.Contains(body, "/signup?email=missing%40example.com") {
		t.Fatalf("body missing signup link: %s", body)
	}
	if sent.link != "" {
		t.Fatalf("sent link = %q, want none", sent.link)
	}
}

func TestLoginMagicLinkSignsInOnce(t *testing.T) {
	sent := &captureMailer{}
	handler, db := newAuthHandler(t, sent)
	router := testRouter(handler)
	createAuthUser(t, db, "ada@example.com")

	body := requestHTML(t, router, http.MethodPost, "/login", url.Values{"email": {"Ada@Example.com"}}, nil)
	if !strings.Contains(body, "Check your email") || !strings.Contains(body, "ada@example.com") || strings.Contains(body, "token=") {
		t.Fatalf("login response = %s", body)
	}
	token := tokenFromLink(t, sent.link)

	confirm := requestHTML(t, router, http.MethodGet, "/login/verify?token="+url.QueryEscape(token), nil, nil)
	if !strings.Contains(confirm, "This link signs you in.") || !strings.Contains(confirm, "Continue") {
		t.Fatalf("confirm page = %s", confirm)
	}

	rec := request(t, router, http.MethodPost, "/login/verify", url.Values{"token": {token}}, nil)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("verify status = %d location = %q", rec.Code, rec.Header().Get("Location"))
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access_token" || !cookies[0].HttpOnly {
		t.Fatalf("cookies = %+v", cookies)
	}

	home := request(t, router, http.MethodGet, "/", nil, cookies)
	if home.Code != http.StatusOK || home.Body.String() != "ok" {
		t.Fatalf("home status = %d body = %s", home.Code, home.Body.String())
	}

	again := requestHTML(t, router, http.MethodPost, "/login/verify", url.Values{"token": {token}}, nil)
	if !strings.Contains(again, "invalid or has expired") {
		t.Fatalf("second verify = %s", again)
	}
}

func TestNewLoginLinkInvalidatesThePreviousOne(t *testing.T) {
	sent := &captureMailer{}
	handler, db := newAuthHandler(t, sent)
	router := testRouter(handler)
	createAuthUser(t, db, "ada@example.com")

	requestHTML(t, router, http.MethodPost, "/login", url.Values{"email": {"ada@example.com"}}, nil)
	first := tokenFromLink(t, sent.link)
	requestHTML(t, router, http.MethodPost, "/login", url.Values{"email": {"ada@example.com"}}, nil)
	second := tokenFromLink(t, sent.link)
	if first == second {
		t.Fatal("expected a new token")
	}

	old := requestHTML(t, router, http.MethodGet, "/login/verify?token="+url.QueryEscape(first), nil, nil)
	if !strings.Contains(old, "invalid or has expired") {
		t.Fatalf("old link = %s", old)
	}
	fresh := requestHTML(t, router, http.MethodGet, "/login/verify?token="+url.QueryEscape(second), nil, nil)
	if !strings.Contains(fresh, "Continue") {
		t.Fatalf("new link = %s", fresh)
	}
}

func TestExpiredMagicLinkIsRejected(t *testing.T) {
	sent := &captureMailer{}
	handler, db := newAuthHandler(t, sent)
	router := testRouter(handler)
	createAuthUser(t, db, "ada@example.com")
	requestHTML(t, router, http.MethodPost, "/login", url.Values{"email": {"ada@example.com"}}, nil)

	_, err := db.Exec(`UPDATE magic_link SET expires_at = ?`, time.Now().UTC().Add(-time.Minute).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("expire link: %v", err)
	}
	body := requestHTML(t, router, http.MethodGet, "/login/verify?token="+url.QueryEscape(tokenFromLink(t, sent.link)), nil, nil)
	if !strings.Contains(body, "invalid or has expired") {
		t.Fatalf("expired link = %s", body)
	}
}

func TestSignupMagicLinkCreatesTheAccount(t *testing.T) {
	sent := &captureMailer{}
	handler, db := newAuthHandler(t, sent)
	router := testRouter(handler)
	users := user.NewUserRepository(db)
	ctx := context.Background()

	body := requestHTML(t, router, http.MethodPost, "/signup", url.Values{"email": {"new@example.com"}}, nil)
	if !strings.Contains(body, "create your account") {
		t.Fatalf("signup response = %s", body)
	}
	_, err := users.GetUserByEmail(ctx, "new@example.com")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("user error before verify = %v, want sql.ErrNoRows", err)
	}

	rec := request(t, router, http.MethodPost, "/login/verify", url.Values{"token": {tokenFromLink(t, sent.link)}}, nil)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("verify status = %d body = %s", rec.Code, rec.Body.String())
	}
	created, err := users.GetUserByEmail(ctx, "new@example.com")
	if err != nil {
		t.Fatalf("get created user: %v", err)
	}
	if created.Email != "new@example.com" || created.FirstName != "" || created.LastName != "" {
		t.Fatalf("created user = %+v", created)
	}
	byID, err := users.GetUserById(ctx, created.ID)
	if err != nil {
		t.Fatalf("get created user by id: %v", err)
	}
	if byID.DisplayName != "" || byID.ProfileID == "" {
		t.Fatalf("created profile = %+v", byID)
	}

	home := request(t, router, http.MethodGet, "/", nil, rec.Result().Cookies())
	if home.Code != http.StatusOK {
		t.Fatalf("home status = %d body = %s", home.Code, home.Body.String())
	}
}

func TestTestMagicLinkEndpointReturnsTheSavedToken(t *testing.T) {
	tokens := mailer.NewTestMailer()
	handler, db := newAuthHandler(t, tokens)
	router := testRouter(handler)
	router.GET("/test/magic-link", auth.TestMagicLink(tokens))
	createAuthUser(t, db, "ada@example.com")

	requestHTML(t, router, http.MethodPost, "/login", url.Values{"email": {"Ada@Example.com"}}, nil)

	missing := request(t, router, http.MethodGet, "/test/magic-link", nil, nil)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing email status = %d body = %s", missing.Code, missing.Body.String())
	}
	unknown := request(t, router, http.MethodGet, "/test/magic-link?email=missing@example.com", nil, nil)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown email status = %d body = %s", unknown.Code, unknown.Body.String())
	}

	rec := request(t, router, http.MethodGet, "/test/magic-link?email=ada@example.com", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("token status = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("token payload = %s", rec.Body.String())
	}

	verified := request(t, router, http.MethodPost, "/login/verify", url.Values{"token": {payload.Token}}, nil)
	if verified.Code != http.StatusSeeOther {
		t.Fatalf("verify status = %d body = %s", verified.Code, verified.Body.String())
	}
}

func TestSignupRejectsAnExistingEmail(t *testing.T) {
	sent := &captureMailer{}
	handler, db := newAuthHandler(t, sent)
	router := testRouter(handler)
	createAuthUser(t, db, "ada@example.com")

	body := requestHTML(t, router, http.MethodPost, "/signup", url.Values{"email": {"ada@example.com"}}, nil)
	if !strings.Contains(body, "already exists") || sent.link != "" {
		t.Fatalf("body = %s link = %q", body, sent.link)
	}
}

func TestLoginRejectsAnInvalidEmail(t *testing.T) {
	sent := &captureMailer{}
	body := requestHTML(t, newTestRouter(t, sent), http.MethodPost, "/login", url.Values{"email": {"not-an-email"}}, nil)
	if !strings.Contains(body, "Enter a valid email address.") || sent.link != "" {
		t.Fatalf("body = %s link = %q", body, sent.link)
	}
}

type captureMailer struct {
	to   string
	link string
	err  error
}

func (m *captureMailer) SendMagicLink(ctx context.Context, to, link string) error {
	m.to = to
	m.link = link
	return m.err
}

func newAuthHandler(t *testing.T, mail mailer.Mailer) (*auth.AuthHandler, *sql.DB) {
	t.Helper()
	db := testdb.Open(t)
	return auth.NewAuthHandler(
		user.NewUserRepository(db),
		auth.NewMagicLinkRepository(db),
		mail,
		sqlite.NewTransactor(db),
	), db
}

func newTestRouter(t *testing.T, mail mailer.Mailer) *gin.Engine {
	t.Helper()
	handler, _ := newAuthHandler(t, mail)
	return testRouter(handler)
}

func testRouter(handler *auth.AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	router.GET("/login", handler.ShowLogin)
	router.POST("/login", handler.RequestLogin)
	router.GET("/signup", handler.ShowSignup)
	router.POST("/signup", handler.RequestSignup)
	router.GET("/login/verify", handler.ShowVerify)
	router.POST("/login/verify", handler.VerifyMagicLink)
	router.GET("/", handler.CheckAccessToken, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return router
}

func createAuthUser(t *testing.T, db *sql.DB, email string) {
	t.Helper()
	repo := user.NewUserRepository(db)
	if _, err := repo.CreateUser(context.Background(), entities.CreateUserParams{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       email,
		ProfileID:   "ada",
		DisplayName: "Ada",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func requestHTML(t *testing.T, router *gin.Engine, method, path string, form url.Values, cookies []*http.Cookie) string {
	t.Helper()
	rec := request(t, router, method, path, form, cookies)
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}

func request(t *testing.T, router *gin.Engine, method, path string, form url.Values, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func tokenFromLink(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse link: %v", err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatalf("link %q has no token", link)
	}
	return token
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
