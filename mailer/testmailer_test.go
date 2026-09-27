package mailer_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"muttley/mailer"
)

func TestTestMailerLogsTheLink(t *testing.T) {
	var buf bytes.Buffer
	m := &mailer.TestMailer{Logger: log.New(&buf, "", 0)}
	link := "http://localhost/login/verify?token=abc"

	if err := m.SendMagicLink(context.Background(), "ada@example.com", link); err != nil {
		t.Fatalf("send magic link: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "ada@example.com") || !strings.Contains(got, link) {
		t.Fatalf("log = %q", got)
	}
}
