package auth

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	purposeLogin        = "login"
	purposeSignup       = "signup"
	magicLinkTTL        = 15 * time.Minute
	magicLinkTokenBytes = 32
)

var errInvalidLink = errors.New("invalid magic link")

func newMagicToken() (raw string, hash string, err error) {
	buf := make([]byte, magicLinkTokenBytes)
	if _, err = cryptorand.Read(buf); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashMagicToken(raw), nil
}

func hashMagicToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(raw string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > 255 || strings.Contains(email, " ") {
		return "", false
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 || !strings.Contains(email[at+1:], ".") {
		return "", false
	}
	return email, true
}

func profileIDForEmail(email string) string {
	local := email
	if at := strings.Index(email, "@"); at > 0 {
		local = email[:at]
	}
	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	name := b.String()
	if name == "" {
		name = "racer"
	}
	if len(name) > 30 {
		name = name[:30]
	}
	return name + "-" + strconv.Itoa(rand.Intn(100000))
}

func signupLink(email string) string {
	if email == "" {
		return "/signup"
	}
	return "/signup?email=" + url.QueryEscape(email)
}

func magicLinkURL(c *gin.Context, token string) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/login/verify?token=" + url.QueryEscape(token)
}
