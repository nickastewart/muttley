package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"muttley/repository"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"
	"strconv"
)

type FriendHandler struct {
	FriendRepository repository.FriendRepository
	UserRepository   repository.UserRepository
}

func NewFriendHandler(friendRepository repository.FriendRepository, userRepository repository.UserRepository) *FriendHandler {
	return &FriendHandler{
		FriendRepository: friendRepository,
		UserRepository:   userRepository,
	}
}

func (controller *FriendHandler) Friends(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	friends, err := controller.FriendRepository.GetFriendsByUser(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting friends"})
		return
	}
	c.HTML(http.StatusOK, "", templates.Friend(friends))
}

func (controller *FriendHandler) SearchFriends(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	searchTerm, ok := c.GetQuery("searchTerm")
	log.Print(searchTerm)
	if !ok {
		log.Panic("Search Term is empty")
	}

	user := u.(entities.GetUserByIdRow)
	searchParams := entities.GetUsersBySearchTermParams{
		Name:   "%" + searchTerm + "%",
		Userid: user.ID,
	}

	results, err := controller.UserRepository.GetUsersBySearchTerm(ctx, searchParams)
	if err != nil {
		log.Panic("Error searching")
	}
	c.HTML(http.StatusOK, "", templates.FriendSearchResults(results))
}

func (controller *FriendHandler) AddFriend(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	friendId, ok := c.GetQuery("userID")

	if !ok {
		log.Panic("Request has no userID")
	}

	friendIdInt, err := strconv.ParseInt(friendId, 10, 63)
	if err != nil {
		log.Panic("Cannot parse friendId")
	}

	addFriendParams := &entities.AddFriendParams{
		UserID:       user.ID,
		FriendID:     friendIdInt,
		FriendStatus: "REQUESTED",
	}

	controller.FriendRepository.AddFriend(ctx, *addFriendParams)
}
