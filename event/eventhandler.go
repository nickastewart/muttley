package event

import (
	"context"
	"log"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/templates"
	"muttley/user"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type EventsHandler struct {
	UserRepository         user.UserRepository
	LocationRepository     location.LocationRepository
	EventRepository        EventRepository
	EventResultRespository eventresult.EventResultRepository
	FriendRepository       friend.FriendRepository
}

func NewEventsHandler(userRepository user.UserRepository,
	eventRepository EventRepository,
	locationRepository location.LocationRepository,
	eventResultRespository eventresult.EventResultRepository,
	friendRepository friend.FriendRepository) *EventsHandler {
	return &EventsHandler{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
		FriendRepository:       friendRepository,
	}
}

func (handler *EventsHandler) Leaderboard(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	userIds := []int64{}
	userIds = append(userIds, user.ID)
	userIds = append(userIds, handler.getFriendIds(ctx, user.ID)...)

	events, err := handler.EventRepository.GetEventsByUser(ctx, userIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].EventResult.BestLapTime < events[j].EventResult.BestLapTime
	})

	c.Header("HX-Redirect", "/leaderboard")
	c.HTML(http.StatusOK, "", templates.Leaderboard(events))
}

func (handler *EventsHandler) getFriendIds(ctx context.Context, userId int64) []int64 {
	friends, err := handler.FriendRepository.GetFriendsByUser(ctx, userId)
	if err != nil {
		log.Fatal("Error getting friends user ids")
	}
	friendIds := make([]int64, len(friends))
	for index, friend := range friends {
		friendIds[index] = friend.ID
	}
	return friendIds
}
