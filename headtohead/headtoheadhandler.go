package headtohead

import (
	"context"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HeadToHeadHandler struct {
	HeadToHeadRepository HeadToHeadRepository
}

func NewHeadToHeadHandler(headToHeadRepository HeadToHeadRepository) *HeadToHeadHandler {
	return &HeadToHeadHandler{
		HeadToHeadRepository: headToHeadRepository,
	}
}

func (handler *HeadToHeadHandler) HeadToHead(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.User)
	results, err := handler.HeadToHeadRepository.GetHeadToHead(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("HX-Redirect", "/head-to-head")
	c.HTML(http.StatusOK, "", templates.HeadToHead(results))
}
