package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	accessTokenCookie = "access_token"
	sessionTTL        = 7 * 24 * time.Hour
)

func CookieSecureFromEnv() (bool, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE"))) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "":
		return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production"), nil
	default:
		return false, fmt.Errorf("auth: unknown COOKIE_SECURE %q (use true or false)", os.Getenv("COOKIE_SECURE"))
	}
}

func ClearSession(c *gin.Context) error {
	return setAccessTokenCookie(c, "", time.Unix(0, 0), -1)
}

func setAccessTokenCookie(c *gin.Context, value string, expires time.Time, maxAge int) error {
	secure, err := CookieSecureFromEnv()
	if err != nil {
		return err
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
