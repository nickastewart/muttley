package mailer

import (
	"fmt"
	"os"
	"strings"
)

// NewFromEnv selects the magic-link mailer.
//
// An empty MAILER or MAILER=log prints the link instead of sending it,
// and is allowed only when APP_ENV=development. Local development should
// set APP_ENV=development and leave MAILER unset. Any other APP_ENV,
// including unset, returns an error. Set MAILER=resend, along with
// RESEND_API_KEY and RESEND_FROM, to deliver mail through Resend.
// MAILER=resend does not depend on APP_ENV.
func NewFromEnv() (Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MAILER"))) {
	case "", "log":
		if strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) != "development" {
			return nil, fmt.Errorf("mailer: log mailer requires APP_ENV=development")
		}
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
