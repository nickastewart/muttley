package mailer_test

import (
	"testing"

	"muttley/mailer"
)

func TestNewFromEnvDefaultsToLogMailer(t *testing.T) {
	t.Setenv("PROFILE", "")
	t.Setenv("MAILER", "")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.LogMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvTestMailer(t *testing.T) {
	t.Setenv("PROFILE", "")
	t.Setenv("MAILER", "test")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.TestMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvTestProfileUsesTestMailer(t *testing.T) {
	t.Setenv("PROFILE", "test")
	t.Setenv("MAILER", "")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.TestMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvResend(t *testing.T) {
	t.Setenv("PROFILE", "")
	t.Setenv("MAILER", "resend")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "login@example.com")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.ResendMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvResendRequiresCredentials(t *testing.T) {
	t.Setenv("PROFILE", "")
	t.Setenv("MAILER", "resend")
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("RESEND_FROM", "login@example.com")
	if _, err := mailer.NewFromEnv(); err == nil {
		t.Fatal("expected missing api key to fail")
	}
}

func TestNewFromEnvRejectsUnknownDriver(t *testing.T) {
	t.Setenv("PROFILE", "")
	t.Setenv("MAILER", "smtp")
	if _, err := mailer.NewFromEnv(); err == nil {
		t.Fatal("expected unknown driver to fail")
	}
}
