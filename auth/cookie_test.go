package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"muttley/auth"
	"muttley/sqlite/entities"
	"muttley/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func TestCookieSecureFromEnv(t *testing.T) {
	cases := []struct {
		name   string
		cookie string
		appEnv string
		want   bool
		fail   bool
	}{
		{name: "true", cookie: "true", appEnv: "development", want: true},
		{name: "true wins over production unset", cookie: " TRUE ", appEnv: "", want: true},
		{name: "false", cookie: "false", appEnv: "production", want: false},
		{name: "false wins over production", cookie: " False ", appEnv: " Production ", want: false},
		{name: "unset production", cookie: "", appEnv: "production", want: true},
		{name: "unset production spaced", cookie: " ", appEnv: " Production ", want: true},
		{name: "unset otherwise", cookie: "", appEnv: "", want: false},
		{name: "unset development", cookie: "", appEnv: "development", want: false},
		{name: "unknown", cookie: "yes", appEnv: "production", fail: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COOKIE_SECURE", tc.cookie)
			t.Setenv("APP_ENV", tc.appEnv)
			got, err := auth.CookieSecureFromEnv()
			if tc.fail {
				if err == nil {
					t.Fatal("expected unknown COOKIE_SECURE to fail")
				}
				return
			}
			if err != nil {
				t.Fatalf("cookie secure: %v", err)
			}
			if got != tc.want {
				t.Fatalf("secure = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCheckAccessTokenMissingCookieDoesNotCallNext(t *testing.T) {
	handler, _ := newAuthHandler(t, &captureMailer{})
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			rec := requestProtected(t, handler, method, nil, nil)
			if rec.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want 303", rec.Code)
			}
			if rec.Header().Get("Location") != "/login" {
				t.Fatalf("location = %q", rec.Header().Get("Location"))
			}
			if strings.Contains(rec.Body.String(), "next") {
				t.Fatalf("body = %s", rec.Body.String())
			}
		})
	}
}

func TestCheckAccessTokenBadClaimsDoNotPanic(t *testing.T) {
	handler, _ := newAuthHandler(t, &captureMailer{})
	future := time.Now().Add(time.Hour).Unix()
	cases := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{name: "missing exp", claims: jwt.MapClaims{"id": float64(1)}},
		{name: "missing id", claims: jwt.MapClaims{"exp": future}},
		{name: "non-float id", claims: jwt.MapClaims{"exp": future, "id": "nope"}},
		{name: "id zero", claims: jwt.MapClaims{"exp": future, "id": float64(0)}},
		{name: "expired", claims: jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix(), "id": float64(1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := requestProtected(t, handler, http.MethodGet, []*http.Cookie{{
				Name:  "access_token",
				Value: signToken(t, tc.claims),
			}}, nil)
			if rec.Code == http.StatusInternalServerError {
				t.Fatalf("status = 500 body = %s", rec.Body.String())
			}
			if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
				t.Fatalf("status = %d location = %q body = %s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), "next") {
				t.Fatalf("body = %s", rec.Body.String())
			}
		})
	}
}

func TestCheckAccessTokenJSONAndHTMX(t *testing.T) {
	handler, _ := newAuthHandler(t, &captureMailer{})

	jsonRec := requestProtected(t, handler, http.MethodPost, nil, map[string]string{
		"Accept": "application/json",
	})
	if jsonRec.Code != http.StatusUnauthorized || strings.Contains(jsonRec.Body.String(), "next") {
		t.Fatalf("json status = %d body = %s", jsonRec.Code, jsonRec.Body.String())
	}
	if !strings.Contains(jsonRec.Body.String(), "Invalid token") {
		t.Fatalf("json body = %s", jsonRec.Body.String())
	}

	both := requestProtected(t, handler, http.MethodGet, nil, map[string]string{
		"Accept": "application/json, text/html",
	})
	if both.Code != http.StatusSeeOther || both.Header().Get("Location") != "/login" {
		t.Fatalf("json+html status = %d location = %q body = %s", both.Code, both.Header().Get("Location"), both.Body.String())
	}

	htmx := requestProtected(t, handler, http.MethodPost, nil, map[string]string{
		"HX-Request": "true",
	})
	if htmx.Code != http.StatusSeeOther || htmx.Header().Get("Location") != "/login" || htmx.Header().Get("HX-Redirect") != "/login" {
		t.Fatalf("htmx status = %d location = %q hx = %q body = %s", htmx.Code, htmx.Header().Get("Location"), htmx.Header().Get("HX-Redirect"), htmx.Body.String())
	}
	if strings.Contains(htmx.Body.String(), "next") {
		t.Fatalf("htmx body = %s", htmx.Body.String())
	}
}

func TestSessionCookieLogoutAndDeleteShareSecure(t *testing.T) {
	cases := []struct {
		name   string
		cookie string
		appEnv string
		want   bool
	}{
		{name: "true", cookie: "true", appEnv: "", want: true},
		{name: "false", cookie: "false", appEnv: "production", want: false},
		{name: "unset production", cookie: "", appEnv: "production", want: true},
		{name: "unset otherwise", cookie: "", appEnv: "development", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COOKIE_SECURE", tc.cookie)
			t.Setenv("APP_ENV", tc.appEnv)

			sent := &captureMailer{}
			handler, db := newAuthHandler(t, sent)
			createAuthUser(t, db, "ada@example.com")
			id := authUserID(t, db, "ada@example.com")

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.HTMLRender = &testTemplRender{}
			router.POST("/login", handler.RequestLogin)
			router.POST("/login/verify", handler.VerifyMagicLink)
			router.POST("/logout", handler.Logout)
			accounts := user.NewUserHandler(user.NewUserRepository(db))
			accounts.ClearSession = auth.ClearSession
			router.POST("/account/delete", func(c *gin.Context) {
				c.Set("currentUser", entities.User{ID: id, Email: "ada@example.com"})
				accounts.DeleteAccount(c)
			})

			request(t, router, http.MethodPost, "/login", url.Values{"email": {"ada@example.com"}}, nil)
			session := request(t, router, http.MethodPost, "/login/verify", url.Values{"token": {tokenFromLink(t, sent.link)}}, nil)
			if session.Code != http.StatusSeeOther {
				t.Fatalf("login status = %d body = %s", session.Code, session.Body.String())
			}
			sessionCookie := accessTokenSetCookie(t, session)
			if secureCookie(sessionCookie) != tc.want {
				t.Fatalf("session cookie = %s", sessionCookie)
			}
			if !strings.Contains(sessionCookie, "; Max-Age=604800") || !strings.Contains(sessionCookie, "; HttpOnly") || !strings.Contains(sessionCookie, "; Path=/") || !strings.Contains(sessionCookie, "; SameSite=Lax") {
				t.Fatalf("session cookie = %s", sessionCookie)
			}
			exp := sessionExp(t, sessionCookie)
			low := time.Now().Add(sessionWeek - time.Minute).Unix()
			high := time.Now().Add(sessionWeek + time.Minute).Unix()
			if exp < low || exp > high {
				t.Fatalf("exp = %d, want between %d and %d", exp, low, high)
			}

			loggedOut := request(t, router, http.MethodPost, "/logout", nil, nil)
			cleared := accessTokenSetCookie(t, loggedOut)
			if secureCookie(cleared) != tc.want || !strings.Contains(cleared, "; Max-Age=0") || !strings.Contains(cleared, "; HttpOnly") || !strings.Contains(cleared, "; Path=/") || !strings.Contains(cleared, "; SameSite=Lax") {
				t.Fatalf("logout cookie = %s", cleared)
			}

			deleted := request(t, router, http.MethodPost, "/account/delete", nil, nil)
			deletedCookie := accessTokenSetCookie(t, deleted)
			if secureCookie(deletedCookie) != tc.want || !strings.Contains(deletedCookie, "; Max-Age=0") || !strings.Contains(deletedCookie, "; SameSite=Lax") {
				t.Fatalf("delete cookie = %s", deletedCookie)
			}
		})
	}
}

const sessionWeek = 7 * 24 * time.Hour

func requestProtected(t *testing.T, handler *auth.AuthHandler, method string, cookies []*http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Handle(method, "/protected", handler.CheckAccessToken, func(c *gin.Context) {
		c.String(http.StatusBadRequest, "next")
	})
	req := httptest.NewRequest(method, "/protected", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func signToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("SECRET"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func accessTokenSetCookie(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var found string
	n := 0
	for _, header := range rec.Header().Values("Set-Cookie") {
		if strings.HasPrefix(header, "access_token=") {
			found = header
			n++
		}
	}
	if n != 1 {
		t.Fatalf("Set-Cookie = %v", rec.Header().Values("Set-Cookie"))
	}
	return found
}

func secureCookie(header string) bool {
	for _, part := range strings.Split(header, ";")[1:] {
		if strings.EqualFold(strings.TrimSpace(part), "Secure") {
			return true
		}
	}
	return false
}

func sessionExp(t *testing.T, header string) int64 {
	t.Helper()
	value := strings.TrimPrefix(header, "access_token=")
	if i := strings.Index(value, ";"); i >= 0 {
		value = value[:i]
	}
	token, _, err := jwt.NewParser().ParseUnverified(value, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims = %T", token.Claims)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp = %#v", claims["exp"])
	}
	return int64(exp)
}
