package mailer

import (
	"context"
	"log"
)

type Mailer interface {
	SendMagicLink(ctx context.Context, to, link string) error
}

type LogMailer struct {
	Logger *log.Logger
}

func NewLogMailer() *LogMailer {
	return &LogMailer{Logger: log.Default()}
}

func (m *LogMailer) SendMagicLink(ctx context.Context, to, link string) error {
	logger := m.Logger
	if logger == nil {
		logger = log.Default()
	}
	logger.Printf("magic link for %s: %s", to, link)
	return nil
}
