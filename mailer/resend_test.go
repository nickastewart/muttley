package mailer_test

import (
	"context"
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"muttley/mailer"
)

func TestResendMailerSendsMagicLink(t *testing.T) {
	var gotAuth, gotType string
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/emails" {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		gotType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	t.Cleanup(srv.Close)

	m, err := mailer.NewResendMailer(mailer.ResendConfig{
		APIKey:  "re_test",
		From:    "Muttley <login@example.com>",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("new resend mailer: %v", err)
	}

	link := `http://localhost/login/verify?token=a&b="c"`
	if err := m.SendMagicLink(context.Background(), "ada@example.com", link); err != nil {
		t.Fatalf("send magic link: %v", err)
	}
	if gotAuth != "Bearer re_test" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if gotType != "application/json" {
		t.Fatalf("content-type = %q", gotType)
	}
	if payload["from"] != "Muttley <login@example.com>" {
		t.Fatalf("from = %#v", payload["from"])
	}
	if payload["subject"] != "Your Muttley sign-in link" {
		t.Fatalf("subject = %#v", payload["subject"])
	}
	to, _ := payload["to"].([]any)
	if len(to) != 1 || to[0] != "ada@example.com" {
		t.Fatalf("to = %#v", payload["to"])
	}
	text, _ := payload["text"].(string)
	htmlBody, _ := payload["html"].(string)
	if !strings.Contains(text, link) {
		t.Fatalf("text = %q", text)
	}
	if !strings.Contains(htmlBody, html.EscapeString(link)) {
		t.Fatalf("html = %q", htmlBody)
	}
}

func TestResendMailerReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"API key is invalid"}`, http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	m, err := mailer.NewResendMailer(mailer.ResendConfig{
		APIKey:  "re_test",
		From:    "login@example.com",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("new resend mailer: %v", err)
	}
	err = m.SendMagicLink(context.Background(), "ada@example.com", "http://localhost/login/verify?token=abc")
	if err == nil || !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "API key is invalid") {
		t.Fatalf("error = %v", err)
	}
}

func TestResendMailerHonorsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(200 * time.Millisecond):
		}
	}))
	t.Cleanup(srv.Close)

	m, err := mailer.NewResendMailer(mailer.ResendConfig{
		APIKey:  "re_test",
		From:    "login@example.com",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("new resend mailer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := m.SendMagicLink(ctx, "ada@example.com", "http://localhost/login/verify?token=abc"); err == nil {
		t.Fatal("expected canceled request to fail")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("canceled request took %s", time.Since(start))
	}
}

func TestResendMailerRejectsBadConfig(t *testing.T) {
	if _, err := mailer.NewResendMailer(mailer.ResendConfig{From: "login@example.com"}); err == nil {
		t.Fatal("expected missing api key to fail")
	}
	if _, err := mailer.NewResendMailer(mailer.ResendConfig{APIKey: "re_test", From: "not-an-email"}); err == nil {
		t.Fatal("expected invalid from address to fail")
	}
}

func TestResendMailerRejectsEmptyLink(t *testing.T) {
	m, err := mailer.NewResendMailer(mailer.ResendConfig{
		APIKey: "re_test",
		From:   "login@example.com",
	})
	if err != nil {
		t.Fatalf("new resend mailer: %v", err)
	}
	if err := m.SendMagicLink(context.Background(), " ", "http://localhost/x"); err == nil {
		t.Fatal("expected empty recipient to fail")
	}
}
