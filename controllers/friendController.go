package controllers

import (
	"context"
	"log"
	"muttley/repository"
	"muttley/sqlite/entities"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FriendController struct {
	FriendRepository repository.FriendRepository
}

func NewFriendController(friendRepository repository.FriendRepository) *FriendController {
	return &FriendController{
		FriendRepository: friendRepository,
	}
}

func (controller *FriendController) AddFriend(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	friendId, ok := c.GetQuery("userID")

	if !ok {
		log.Fatal("Request has no userID")
	}

	friendIdInt, err := strconv.ParseInt(friendId, 10, 63)
	if err != nil {
		log.Fatal("Cannot parse friendId")
	}

	addFriendParams := &entities.AddFriendParams{
		UserID:       user.ID,
		FriendID:     friendIdInt,
		FriendStatus: "REQUESTED",
	}

	controller.FriendRepository.AddFriend(ctx, *addFriendParams)
}
