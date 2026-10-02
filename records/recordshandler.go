package records

import (
	"context"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecordsHandler struct {
	RecordsRepository RecordsRepository
}

func NewRecordsHandler(recordsRepository RecordsRepository) *RecordsHandler {
	return &RecordsHandler{
		RecordsRepository: recordsRepository,
	}
}

func (handler *RecordsHandler) Records(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.User)
	rows, err := handler.RecordsRepository.ListLocationRecords(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("HX-Redirect", "/records")
	c.HTML(http.StatusOK, "", templates.Records(rows))
}
