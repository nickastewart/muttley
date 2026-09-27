package mailer

import (
	"fmt"
	"os"
	"strings"
)

// NewFromEnv selects the magic-link mailer.
//
// MAILER defaults to log, which prints the link instead of sending it.
// Local development should leave MAILER unset. Set MAILER=resend, along
// with RESEND_API_KEY and RESEND_FROM, to deliver mail through Resend.
func NewFromEnv() (Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MAILER"))) {
	case "", "log":
		return NewLogMailer(), nil
	case "resend":
		return NewResendMailer(ResendConfig{
			APIKey: os.Getenv("RESEND_API_KEY"),
			From:   os.Getenv("RESEND_FROM"),
		})
	default:
		return nil, fmt.Errorf("mailer: unknown MAILER %q (use log or resend)", os.Getenv("MAILER"))
	}
}
