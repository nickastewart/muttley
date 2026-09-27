package auth

import (
	"net/http"

	"muttley/mailer"

	"github.com/gin-gonic/gin"
)

// TestMagicLink returns the raw token TestMailer saved for an email.
// Register it only when MAILER=test. The token signs the user in.
func TestMagicLink(tokens *mailer.TestMailer) gin.HandlerFunc {
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
		saved, ok := tokens.Token(email)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "no magic link token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": saved.Token})
	}
}
