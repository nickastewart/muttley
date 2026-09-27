package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"muttley/mailer"

	"github.com/gin-gonic/gin"
)

type magicLinkTokenReader interface {
	ActiveToken(ctx context.Context, email string) (string, error)
}

// RegisterTestMagicLink adds GET /test/magic-link/:email when PROFILE=test.
// The route is left unregistered for every other profile.
func RegisterTestMagicLink(router gin.IRoutes, tokens magicLinkTokenReader) {
	if !mailer.TestProfile() {
		return
	}
	router.GET("/test/magic-link/:email", TestMagicLink(tokens))
}

// TestMagicLink returns the raw token stored for the email path parameter.
// It responds 404 unless PROFILE=test. The token signs the user in.
func TestMagicLink(tokens magicLinkTokenReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mailer.TestProfile() || tokens == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		email, ok := normalizeEmail(c.Param("email"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
			return
		}
		token, err := tokens.ActiveToken(c.Request.Context(), email)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no magic link token"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load magic link token"})
			return
		}
		if token == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no magic link token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
