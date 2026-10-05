package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRedactTokenQuery(t *testing.T) {
	const secret = "SECRET"
	withToken := redactTokenQuery("/login/verify?token=" + secret)
	if strings.Contains(withToken, secret) || !strings.Contains(withToken, "/login/verify") || !strings.Contains(withToken, "token=REDACTED") {
		t.Fatalf("redacted = %q", withToken)
	}

	unchanged := []string{
		"/login/verify",
		"/dashboard",
		"/login/verify?email=ada%40example.com",
		"/login/verify?next=token",
	}
	for _, path := range unchanged {
		if got := redactTokenQuery(path); got != path {
			t.Fatalf("redact(%q) = %q", path, got)
		}
	}

	beside := []string{
		"/login/verify?email=ada%40example.com&token=" + secret,
		"/login/verify?token=" + secret + "&email=ada%40example.com",
	}
	for _, path := range beside {
		got := redactTokenQuery(path)
		if strings.Contains(got, secret) || !strings.Contains(got, "/login/verify") || !strings.Contains(got, "email=ada%40example.com") || !strings.Contains(got, "token=REDACTED") {
			t.Fatalf("redact(%q) = %q", path, got)
		}
	}
}

func TestAccessLogRedactsVerifyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	previous := gin.DefaultWriter
	gin.DefaultWriter = &buf
	t.Cleanup(func() {
		gin.DefaultWriter = previous
	})

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(accessLogFormatter), gin.Recovery())
	router.GET("/login/verify", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/login/verify?token=supersecret", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	got := buf.String()
	if !strings.Contains(got, http.MethodGet) || !strings.Contains(got, "/login/verify") || !strings.Contains(got, "token=REDACTED") {
		t.Fatalf("log = %q", got)
	}
	if strings.Contains(got, "supersecret") {
		t.Fatalf("log leaked token: %q", got)
	}
}
