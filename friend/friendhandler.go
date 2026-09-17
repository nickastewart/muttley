package friend

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"muttley/sqlite/entities"
	"muttley/templates"
	"muttley/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	FriendRepository FriendRepository
	UserRepository   user.UserRepository
}

func NewFriendHandler(friendRepository FriendRepository, userRepository user.UserRepository) *FriendHandler {
	return &FriendHandler{
		FriendRepository: friendRepository,
		UserRepository:   userRepository,
	}
}
func (handler *FriendHandler) Friends(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	friends, err := handler.FriendRepository.GetFriendsByUser(ctx, user.ID)
	for _, friend := range friends {
		log.Print(friend)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting friends"})
		return
	}
	c.Header("HX-Redirect", "/friends")
	c.HTML(http.StatusOK, "", templates.Friend(friends))
}

func (handler *FriendHandler) SearchFriends(c *gin.Context) {
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

	results, err := handler.UserRepository.GetUsersBySearchTerm(ctx, searchParams)
	if err != nil {
		log.Panic("Error searching")
	}
	c.HTML(http.StatusOK, "", templates.FriendSearchResults(results))
}

func (handler *FriendHandler) AddFriend(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	profileID, ok := c.GetQuery("profileId")

	if !ok || profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request has no profileID"})
		return
	}

	friendId, err := handler.UserRepository.GetUserIdByProfileId(ctx, profileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot find user"})
		return
	}

	params := entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   user.ID,
		Friendid: friendId,
	}
	friend, err := handler.FriendRepository.GetFriendByUserIdAndFriendId(ctx, params)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error looking up friend"})
		return
	}

	if err == nil && friend.ID != 0 {
		handler.updateFriendStatus(ctx, friend.ID, "REQUESTED")
	} else {
		addFriendParams := entities.AddFriendParams{
			UserID:       user.ID,
			FriendID:     friendId,
			FriendStatus: "REQUESTED",
		}

		_, err = handler.FriendRepository.AddFriend(ctx, addFriendParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding friend"})
			return
		}
	}

	c.HTML(http.StatusOK, "", templates.FriendPendingRequestButton())
}

func (handler *FriendHandler) RemoveFriend(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	profileID, ok := c.GetQuery("profileId")

	if !ok {
		log.Panic("Request has no profileID")
	}

	friendId, err := handler.UserRepository.GetUserIdByProfileId(ctx, profileID)
	if err != nil {
		log.Panic("Cannot parse friendId")
	}

	params := entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   user.ID,
		Friendid: friendId,
	}

	friend, err := handler.FriendRepository.GetFriendByUserIdAndFriendId(ctx, params)
	if err != nil {
		log.Panic("Cannot friend from user and id")
	}

	if friend.FriendStatus != "CANCELLED" {
		handler.updateFriendStatus(ctx, friend.ID, "CANCELLED")
	}
}

func (handler *FriendHandler) updateFriendStatus(ctx context.Context, friendId int64, status string) {

	updateStatusParams := entities.UpdateFriendStatusParams{
		FriendStatus: status,
		ID:           friendId,
	}
	handler.FriendRepository.UpdateFriendStatus(ctx, updateStatusParams)
}
