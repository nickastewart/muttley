package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type magicLinkTokenReader interface {
	ActiveToken(ctx context.Context, email string) (string, error)
}

// TestMagicLink returns the raw token stored for an email.
// Register it only when MAILER=test. The token signs the user in.
func TestMagicLink(tokens magicLinkTokenReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokens == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "magic link test mode is not enabled"})
			return
		}
		email, ok := normalizeEmail(c.Query("email"))
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
