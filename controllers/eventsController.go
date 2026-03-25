package controllers

import (
	"context"
	"log"
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
	FriendRepository       repository.FriendRepository
}

func NewEventsController(userRepository repository.UserRepository,
	eventRepository repository.EventRepository,
	locationRepository repository.LocationRepository,
	eventResultRespository repository.EventResultRepository,
	friendRepository repository.FriendRepository) *EventsController {
	return &EventsController{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
		FriendRepository:       friendRepository,
	}
}

func (controller *EventsController) Leaderboard(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	userIds := []int64{}
	userIds = append(userIds, user.ID)
	userIds = append(userIds, controller.getFriendIds(ctx, user.ID)...)

	events, err := controller.EventRepository.GetEventsByUser(ctx, userIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "", templates.Leaderboard(events))
}

func (controller *EventsController) getFriendIds(ctx context.Context, userId int64) []int64 {
	friends, err := controller.FriendRepository.GetFriendsByUser(ctx, userId)
	if err != nil {
		log.Fatal("Error getting friends user ids")
	}
	friendIds := make([]int64, len(friends))
	for index, friend := range friends {
		friendIds[index] = friend.ID
	}
	return friendIds
}
