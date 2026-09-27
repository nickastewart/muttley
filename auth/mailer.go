package auth

import (
	"log"
)

type Mailer interface {
	SendMagicLink(toEmail string, loginURL string) error
}

// LogMailer writes the magic link to the process log. Use this in local
// development until an SMTP or transactional email provider is wired up.
type LogMailer struct{}

func (LogMailer) SendMagicLink(toEmail string, loginURL string) error {
	log.Printf("magic link for %s: %s", toEmail, loginURL)
	return nil
}
