package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TestProfile reports whether this process is running the test profile.
// Set PROFILE=test. The test login route is registered only then.
func TestProfile() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PROFILE")), "test")
}

// RegisterTestLogin adds GET /test/login/:id when PROFILE=test.
// The route is left unregistered for every other profile.
func RegisterTestLogin(router gin.IRoutes, handler *AuthHandler) {
	if !TestProfile() || handler == nil {
		return
	}
	router.GET("/test/login/:id", handler.TestLogin)
}

// TestLogin sets the access_token cookie for the user id path parameter.
// It responds 404 unless PROFILE=test.
func (handler *AuthHandler) TestLogin(c *gin.Context) {
	if !TestProfile() || handler == nil || handler.UserRepository == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}
	found, err := handler.UserRepository.GetUserById(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && found.ID == 0) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
		return
	}
	if err := handler.setSession(c, found.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not set session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": found.ID})
}
