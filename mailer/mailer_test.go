package mailer_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"muttley/mailer"
)

func TestLogMailerWritesTheLink(t *testing.T) {
	var buf bytes.Buffer
	m := &mailer.LogMailer{Logger: log.New(&buf, "", 0)}

	err := m.SendMagicLink(context.Background(), "ada@example.com", "http://localhost/login/verify?token=abc")
	if err != nil {
		t.Fatalf("send magic link: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "ada@example.com") || !strings.Contains(got, "http://localhost/login/verify?token=abc") {
		t.Fatalf("log = %q", got)
	}
}
