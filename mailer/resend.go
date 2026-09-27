package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultResendBaseURL = "https://api.resend.com"
	magicLinkSubject     = "Your Muttley sign-in link"
)

// ResendMailer sends magic-link email through the Resend HTTP API.
// https://resend.com/docs/api-reference/emails/send-email
type ResendMailer struct {
	apiKey  string
	from    string
	baseURL string
	client  *http.Client
}

type ResendConfig struct {
	APIKey  string
	From    string
	BaseURL string
	Client  *http.Client
}

func NewResendMailer(cfg ResendConfig) (*ResendMailer, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	from := strings.TrimSpace(cfg.From)
	if apiKey == "" {
		return nil, fmt.Errorf("mailer: RESEND_API_KEY is required")
	}
	if from == "" || !strings.Contains(from, "@") {
		return nil, fmt.Errorf("mailer: RESEND_FROM must be an email address")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultResendBaseURL
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ResendMailer{
		apiKey:  apiKey,
		from:    from,
		baseURL: baseURL,
		client:  client,
	}, nil
}

type resendEmail struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

func (m *ResendMailer) SendMagicLink(ctx context.Context, to, link string) error {
	to = strings.TrimSpace(to)
	link = strings.TrimSpace(link)
	if to == "" || link == "" {
		return fmt.Errorf("mailer: magic link recipient and link are required")
	}

	body, err := json.Marshal(resendEmail{
		From:    m.from,
		To:      []string{to},
		Subject: magicLinkSubject,
		HTML:    magicLinkHTML(link),
		Text:    magicLinkText(link),
	})
	if err != nil {
		return fmt.Errorf("mailer: encode resend request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mailer: build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: send resend request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return fmt.Errorf("mailer: read resend response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("mailer: resend returned %d: %s", resp.StatusCode, msg)
	}
	return nil
}

func magicLinkText(link string) string {
	return "Use this link to sign in to Muttley. It expires in 15 minutes.\n\n" + link + "\n"
}

func magicLinkHTML(link string) string {
	escaped := html.EscapeString(link)
	return "<p>Use this link to sign in to Muttley. It expires in 15 minutes.</p>" +
		"<p><a href=\"" + escaped + "\">Sign in to Muttley</a></p>" +
		"<p>" + escaped + "</p>"
}
