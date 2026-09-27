package mailer

import (
	"context"
	"log"
)

// TestMailer logs a magic link. The sign-in record is the token hash already
// stored on magic_link, the same as a real user. It does not keep the raw token.
type TestMailer struct {
	Logger *log.Logger
}

func NewTestMailer() *TestMailer {
	return &TestMailer{Logger: log.Default()}
}

func (m *TestMailer) SendMagicLink(ctx context.Context, to, link string) error {
	logger := m.Logger
	if logger == nil {
		logger = log.Default()
	}
	logger.Printf("magic link for %s: %s", to, link)
	return nil
}
