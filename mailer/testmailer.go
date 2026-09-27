package mailer

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
)

// MagicLinkToken is a raw sign-in token kept by TestMailer so a load test
// can fetch it. The hashed copy used to sign in still lives in magic_link.
type MagicLinkToken struct {
	Email string
	Token string
	Link  string
}

// TestMailer logs a magic link and remembers the raw token, keyed by email.
// It is only for load tests. Do not use it in production: the token endpoint
// returns secrets that sign a user in.
type TestMailer struct {
	Logger *log.Logger

	mu     sync.Mutex
	tokens map[string]MagicLinkToken
}

func NewTestMailer() *TestMailer {
	return &TestMailer{
		Logger: log.Default(),
		tokens: map[string]MagicLinkToken{},
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

	logger := m.Logger
	if logger == nil {
		logger = log.Default()
	}
	logger.Printf("magic link for %s: %s", email, link)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tokens == nil {
		m.tokens = map[string]MagicLinkToken{}
	}
	m.tokens[email] = MagicLinkToken{Email: email, Token: token, Link: link}
	return nil
}

// Token returns the latest raw magic-link token saved for email.
func (m *TestMailer) Token(email string) (MagicLinkToken, bool) {
	key := strings.ToLower(strings.TrimSpace(email))
	m.mu.Lock()
	defer m.mu.Unlock()
	saved, ok := m.tokens[key]
	return saved, ok
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
