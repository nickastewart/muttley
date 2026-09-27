package mailer_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"muttley/mailer"
)

type savedToken struct {
	email string
	token string
}

type fakeTokenStore struct {
	saved []savedToken
	err   error
}

func (s *fakeTokenStore) SaveToken(ctx context.Context, email, token string) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, savedToken{email: email, token: token})
	return nil
}

func TestTestMailerLogsAndSavesTheToken(t *testing.T) {
	var buf bytes.Buffer
	store := &fakeTokenStore{}
	m := &mailer.TestMailer{Logger: log.New(&buf, "", 0), Store: store}
	link := "http://localhost/login/verify?token=abc"

	if err := m.SendMagicLink(context.Background(), "Ada@Example.com", link); err != nil {
		t.Fatalf("send magic link: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "ada@example.com") || !strings.Contains(got, link) {
		t.Fatalf("log = %q", got)
	}
	if len(store.saved) != 1 || store.saved[0].email != "ada@example.com" || store.saved[0].token != "abc" {
		t.Fatalf("saved = %+v", store.saved)
	}
}

func TestTestMailerRequiresAStore(t *testing.T) {
	m := mailer.NewTestMailer(nil)
	err := m.SendMagicLink(context.Background(), "ada@example.com", "http://localhost/login/verify?token=abc")
	if err == nil {
		t.Fatal("expected a missing store to fail")
	}
}

func TestTestMailerRejectsALinkWithoutAToken(t *testing.T) {
	store := &fakeTokenStore{}
	m := mailer.NewTestMailer(store)
	err := m.SendMagicLink(context.Background(), "ada@example.com", "http://localhost/login/verify")
	if err == nil {
		t.Fatal("expected missing token to fail")
	}
	if len(store.saved) != 0 {
		t.Fatalf("saved = %+v", store.saved)
	}
}
