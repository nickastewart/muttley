package mailer_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"muttley/mailer"
)

func TestTestMailerLogsAndSavesTheToken(t *testing.T) {
	var buf bytes.Buffer
	m := &mailer.TestMailer{Logger: log.New(&buf, "", 0)}
	link := "http://localhost/login/verify?token=abc"

	if err := m.SendMagicLink(context.Background(), "Ada@Example.com", link); err != nil {
		t.Fatalf("send magic link: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "ada@example.com") || !strings.Contains(got, link) {
		t.Fatalf("log = %q", got)
	}

	saved, ok := m.Token("ada@example.com")
	if !ok || saved.Token != "abc" || saved.Link != link || saved.Email != "ada@example.com" {
		t.Fatalf("saved = %+v ok = %v", saved, ok)
	}
}

func TestTestMailerKeepsTheLatestTokenPerEmail(t *testing.T) {
	m := mailer.NewTestMailer()
	ctx := context.Background()
	if err := m.SendMagicLink(ctx, "ada@example.com", "http://localhost/login/verify?token=first"); err != nil {
		t.Fatalf("first link: %v", err)
	}
	if err := m.SendMagicLink(ctx, "ada@example.com", "http://localhost/login/verify?token=second"); err != nil {
		t.Fatalf("second link: %v", err)
	}
	if err := m.SendMagicLink(ctx, "grace@example.com", "http://localhost/login/verify?token=other"); err != nil {
		t.Fatalf("other link: %v", err)
	}

	ada, ok := m.Token("ada@example.com")
	if !ok || ada.Token != "second" {
		t.Fatalf("ada = %+v ok = %v", ada, ok)
	}
	grace, ok := m.Token("grace@example.com")
	if !ok || grace.Token != "other" {
		t.Fatalf("grace = %+v ok = %v", grace, ok)
	}
}

func TestTestMailerRejectsALinkWithoutAToken(t *testing.T) {
	m := mailer.NewTestMailer()
	err := m.SendMagicLink(context.Background(), "ada@example.com", "http://localhost/login/verify")
	if err == nil {
		t.Fatal("expected missing token to fail")
	}
	if _, ok := m.Token("ada@example.com"); ok {
		t.Fatal("saved a token without one")
	}
}
