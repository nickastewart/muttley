package mailer_test

import (
	"testing"

	"muttley/mailer"
)

func TestNewFromEnvDefaultsToLogMailer(t *testing.T) {
	t.Setenv("MAILER", "")
	t.Setenv("APP_ENV", "development")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.LogMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvLogMailerAllowsDevelopment(t *testing.T) {
	t.Setenv("MAILER", "log")
	t.Setenv("APP_ENV", " Development ")
	got, err := mailer.NewFromEnv()
	if err != nil {
		t.Fatalf("new from env: %v", err)
	}
	if _, ok := got.(*mailer.LogMailer); !ok {
		t.Fatalf("mailer = %T", got)
	}
}

func TestNewFromEnvLogMailerRequiresDevelopment(t *testing.T) {
	cases := []struct {
		mailer string
		appEnv string
	}{
		{mailer: "", appEnv: ""},
		{mailer: "", appEnv: "production"},
		{mailer: "log", appEnv: ""},
		{mailer: "log", appEnv: "production"},
	}
	for _, tc := range cases {
		t.Run(tc.mailer+"/"+tc.appEnv, func(t *testing.T) {
			t.Setenv("MAILER", tc.mailer)
			t.Setenv("APP_ENV", tc.appEnv)
			if _, err := mailer.NewFromEnv(); err == nil {
				t.Fatal("expected log mailer to be refused")
			}
		})
	}
}

func TestNewFromEnvResend(t *testing.T) {
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

func TestNewFromEnvResendIgnoresAppEnv(t *testing.T) {
	t.Setenv("MAILER", "resend")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "login@example.com")
	for _, appEnv := range []string{"", "production"} {
		t.Run(appEnv, func(t *testing.T) {
			t.Setenv("APP_ENV", appEnv)
			got, err := mailer.NewFromEnv()
			if err != nil {
				t.Fatalf("new from env: %v", err)
			}
			if _, ok := got.(*mailer.ResendMailer); !ok {
				t.Fatalf("mailer = %T", got)
			}
		})
	}
}

func TestNewFromEnvResendRequiresCredentials(t *testing.T) {
	t.Setenv("MAILER", "resend")
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("RESEND_FROM", "login@example.com")
	if _, err := mailer.NewFromEnv(); err == nil {
		t.Fatal("expected missing api key to fail")
	}
}

func TestNewFromEnvRejectsUnknownDriver(t *testing.T) {
	t.Setenv("MAILER", "smtp")
	if _, err := mailer.NewFromEnv(); err == nil {
		t.Fatal("expected unknown driver to fail")
	}
}
