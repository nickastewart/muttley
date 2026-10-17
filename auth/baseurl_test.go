package auth_test

import (
	"strings"
	"testing"

	"muttley/auth"
)

func TestParsePublicBaseURL(t *testing.T) {
	accepted := map[string]string{
		"https://muttley.example":     "https://muttley.example",
		"https://muttley.example/":    "https://muttley.example",
		"https://muttley.example:443": "https://muttley.example:443",
		"http://localhost:8080":       "http://localhost:8080",
		"http://127.0.0.1:9090":       "http://127.0.0.1:9090",
		"http://localhost":            "http://localhost",
	}
	for raw, want := range accepted {
		got, err := auth.ParsePublicBaseURL(raw)
		if err != nil || got != want {
			t.Fatalf("ParsePublicBaseURL(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}

	rejected := []string{
		"http://muttley.example",
		"http://muttley.example:8080",
		"https://user:secret@muttley.example",
		"https://muttley.example/login",
		"https://muttley.example?token=1",
		"https://muttley.example#verify",
		"ftp://muttley.example",
		"muttley.example",
		"",
	}
	for _, raw := range rejected {
		if _, err := auth.ParsePublicBaseURL(raw); err == nil {
			t.Fatalf("ParsePublicBaseURL(%q) accepted, want error", raw)
		}
	}
}

func TestBaseURLFromEnv(t *testing.T) {
	t.Run("log mailer default", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		for _, mailer := range []string{"", "log", " LOG "} {
			t.Setenv("MAILER", mailer)
			got, err := auth.BaseURLFromEnv()
			if err != nil || got != "http://localhost:8080" {
				t.Fatalf("MAILER=%q BaseURLFromEnv() = %q, %v", mailer, got, err)
			}
		}
	})

	t.Run("resend requires the env", func(t *testing.T) {
		t.Setenv("APP_BASE_URL", "")
		t.Setenv("MAILER", "resend")
		if _, err := auth.BaseURLFromEnv(); err == nil || !strings.Contains(err.Error(), "APP_BASE_URL") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("explicit value", func(t *testing.T) {
		t.Setenv("MAILER", "resend")
		t.Setenv("APP_BASE_URL", "https://muttley.example/")
		got, err := auth.BaseURLFromEnv()
		if err != nil || got != "https://muttley.example" {
			t.Fatalf("BaseURLFromEnv() = %q, %v", got, err)
		}
	})
}
