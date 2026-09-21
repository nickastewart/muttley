package user

import (
	"context"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserRepository UserRepository
}

func NewUserHandler(userRepository UserRepository) *UserHandler {
	return &UserHandler{
		UserRepository: userRepository,
	}
}

func (handler *UserHandler) Account(c *gin.Context) {
	u, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	c.Header("HX-Redirect", "/account")
	c.HTML(http.StatusOK, "", templates.Account(user, "", ""))
}

func (handler *UserHandler) UpdateAccount(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	currentEmail := user.Email

	var form UpdateAccountForm
	c.Bind(&form)
	form.FirstName = strings.TrimSpace(form.FirstName)
	form.LastName = strings.TrimSpace(form.LastName)
	form.DisplayName = strings.TrimSpace(form.DisplayName)
	form.Email = strings.TrimSpace(form.Email)

	user.FirstName = form.FirstName
	user.LastName = form.LastName
	user.DisplayName = form.DisplayName
	user.Email = form.Email

	if form.FirstName == "" || form.LastName == "" || form.DisplayName == "" || form.Email == "" {
		c.HTML(http.StatusBadRequest, "", templates.Account(user, "", "Please fill in all fields."))
		return
	}

	if form.Email != currentEmail {
		existing, _ := handler.UserRepository.GetUserByEmail(ctx, form.Email)
		if existing.ID != 0 && existing.ID != user.ID {
			c.HTML(http.StatusBadRequest, "", templates.Account(user, "", "A user with this email already exists."))
			return
		}
	}

	err := handler.UserRepository.UpdateUser(ctx, entities.UpdateUserParams{
		FirstName:   form.FirstName,
		LastName:    form.LastName,
		Email:       form.Email,
		DisplayName: form.DisplayName,
		ID:          user.ID,
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "", templates.Account(user, "", "Unable to update account. Please try again."))
		return
	}

	updated, err := handler.UserRepository.GetUserById(ctx, user.ID)
	if err != nil {
		c.HTML(http.StatusOK, "", templates.Account(user, "Account updated.", ""))
		return
	}

	c.HTML(http.StatusOK, "", templates.Account(updated, "Account updated.", ""))
}

func (handler *UserHandler) DeleteAccount(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	if err := handler.UserRepository.DeleteUser(ctx, user.ID); err != nil {
		c.Header("HX-Retarget", "body")
		c.Header("HX-Reswap", "innerHTML")
		c.HTML(http.StatusBadRequest, "", templates.Account(user, "", "Unable to delete account. Please try again."))
		return
	}

	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.Header("HX-Redirect", "/login")
	if c.GetHeader("HX-Request") != "true" {
		c.Redirect(http.StatusSeeOther, "/login")
	}
}
