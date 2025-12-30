package controllers

import (
	"context"
	"muttley/repository"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EventsController struct {
	UserRepository         repository.UserRepository
	LocationRepository     repository.LocationRepository
	EventRepository        repository.EventRepository
	EventResultRespository repository.EventResultRepository
}

func NewEventsController(userRepository repository.UserRepository,
	eventRepository repository.EventRepository,
	locationRepository repository.LocationRepository,
	eventResultRespository repository.EventResultRepository) *EventsController {
	return &EventsController{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
	}
}

func (controller *EventsController) GetEventsByUser(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	events, err := controller.EventRepository.GetEventsByUser(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "", templates.Leaderboard(events))
}
