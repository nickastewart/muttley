package mailer

import (
	"fmt"
	"os"
	"strings"
)

// TestProfile reports whether this process is running the test profile.
// Set PROFILE=test. The magic-link token endpoint is registered only then.
func TestProfile() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PROFILE")), "test")
}

// NewFromEnv selects the magic-link mailer.
//
// MAILER defaults to log, which prints the link instead of sending it.
// Local development should leave MAILER unset. Set MAILER=resend, along
// with RESEND_API_KEY and RESEND_FROM, to deliver mail through Resend.
// PROFILE=test, or MAILER=test, logs the link and stores the raw token in
// the database. The token endpoint is still registered only for PROFILE=test.
func NewFromEnv() (Mailer, error) {
	if TestProfile() || strings.EqualFold(strings.TrimSpace(os.Getenv("MAILER")), "test") {
		return NewTestMailer(nil), nil
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MAILER"))) {
	case "", "log":
		return NewLogMailer(), nil
	case "resend":
		return NewResendMailer(ResendConfig{
			APIKey: os.Getenv("RESEND_API_KEY"),
			From:   os.Getenv("RESEND_FROM"),
		})
	default:
		return nil, fmt.Errorf("mailer: unknown MAILER %q (use log, test, or resend)", os.Getenv("MAILER"))
	}
}
