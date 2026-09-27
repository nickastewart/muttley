package mailer

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

// TokenStore saves the raw magic-link token for a later database lookup.
type TokenStore interface {
	SaveToken(ctx context.Context, email, token string) error
}

// TestMailer logs a magic link and stores the raw token for that email.
// It is only for load tests. Do not use it in production: the token endpoint
// returns a secret that signs a user in.
type TestMailer struct {
	Logger *log.Logger
	Store  TokenStore
}

func NewTestMailer(store TokenStore) *TestMailer {
	return &TestMailer{
		Logger: log.Default(),
		Store:  store,
	}
}

func (m *TestMailer) SendMagicLink(ctx context.Context, to, link string) error {
	token, err := tokenFromMagicLink(link)
	if err != nil {
		return err
	}
	email := strings.ToLower(strings.TrimSpace(to))
	if email == "" {
		return fmt.Errorf("mailer: magic link email is required")
	}
	if m.Store == nil {
		return fmt.Errorf("mailer: test token store is not configured")
	}

	logger := m.Logger
	if logger == nil {
		logger = log.Default()
	}
	logger.Printf("magic link for %s: %s", email, link)
	return m.Store.SaveToken(ctx, email, token)
}

func tokenFromMagicLink(link string) (string, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("mailer: parse magic link: %w", err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		return "", fmt.Errorf("mailer: magic link has no token")
	}
	return token, nil
}
