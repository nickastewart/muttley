package auth_test

import (
	"strings"
	"testing"

	"muttley/auth"
)

func TestSigningKeyFromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := auth.SigningKeyFromEnv(); err == nil {
		t.Fatal("expected missing JWT_SECRET to fail")
	}

	t.Setenv("JWT_SECRET", strings.Repeat("a", 31))
	if _, err := auth.SigningKeyFromEnv(); err == nil {
		t.Fatal("expected short JWT_SECRET to fail")
	}

	t.Setenv("JWT_SECRET", " "+strings.Repeat("a", 31)+" ")
	if _, err := auth.SigningKeyFromEnv(); err == nil {
		t.Fatal("expected trimmed short JWT_SECRET to fail")
	}

	leaked := "supersecret-but-too-short"
	t.Setenv("JWT_SECRET", leaked)
	_, err := auth.SigningKeyFromEnv()
	if err == nil || strings.Contains(err.Error(), leaked) || strings.Contains(err.Error(), "JWT_SECRET=") {
		t.Fatalf("err = %v", err)
	}

	want := strings.Repeat("b", 32)
	t.Setenv("JWT_SECRET", " "+want+" ")
	got, err := auth.SigningKeyFromEnv()
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	if string(got) != want {
		t.Fatalf("key = %q", got)
	}
}
