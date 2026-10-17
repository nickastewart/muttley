package auth

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func BaseURLFromEnv() (string, error) {
	raw := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	if raw == "" {
		switch strings.ToLower(strings.TrimSpace(os.Getenv("MAILER"))) {
		case "", "log":
			return "http://localhost:8080", nil
		default:
			return "", fmt.Errorf("auth: APP_BASE_URL is required when MAILER=%s", strings.ToLower(strings.TrimSpace(os.Getenv("MAILER"))))
		}
	}
	return ParsePublicBaseURL(raw)
}

func ParsePublicBaseURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" {
		return "", fmt.Errorf("auth: APP_BASE_URL must be an absolute URL with a host")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("auth: APP_BASE_URL must not include user info")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("auth: APP_BASE_URL must not include a query or fragment")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("auth: APP_BASE_URL must not include a path")
	}
	switch parsed.Scheme {
	case "https":
	case "http":
		host := parsed.Hostname()
		if host != "localhost" && host != "127.0.0.1" {
			return "", fmt.Errorf("auth: APP_BASE_URL may use http only for localhost")
		}
	default:
		return "", fmt.Errorf("auth: APP_BASE_URL scheme must be https")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}
