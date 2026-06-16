package auth

import (
	"context"
	"fmt"
	"muttley/sqlite/entities"
	"muttley/user"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
)

type AuthHandler struct {
	UserRepository user.UserRepository
}

func NewAuthHandler(userRepository user.UserRepository) *AuthHandler {
	return &AuthHandler{
		UserRepository: userRepository,
	}
}

func (handler *AuthHandler) Signup(c *gin.Context) {
	ctx := context.Background()

	var signupForm SignupForm
	c.Bind(&signupForm)

	userFound, err := handler.UserRepository.GetUserByEmail(ctx, signupForm.Email)
	if userFound.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signupForm.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createUserParams := entities.CreateUserParams{
		FirstName:   signupForm.FirstName,
		LastName:    signupForm.LastName,
		Email:       signupForm.Email,
		Password:    string(passwordHash),
		ProfileID:   generateProfileId(signupForm.FirstName, signupForm.LastName),
		DisplayName: signupForm.DisplayName,
	}

	_, err = handler.UserRepository.CreateUser(ctx, createUserParams)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process request"})
		return
	}

	c.HTML(http.StatusOK, "", Login(nil))
}

func generateProfileId(firstName string, lastName string) string {
	randomInt := rand.Intn(100000)
	return firstName + "-" + lastName + "-" + strconv.Itoa(randomInt)
}

func (handler *AuthHandler) LoginForm(c *gin.Context) {
	ctx := context.Background()
	var loginForm LoginForm

	c.Bind(&loginForm)

	userFound, _ := handler.UserRepository.GetUserByEmailForLogin(ctx, loginForm.Email)

	if userFound.ID == 0 {
		loginErr := &AuthError{
			Message: "Invalid username or password. Please try again.",
			Type:    "Authentication",
		}
		c.HTML(http.StatusForbidden, "", Login(loginErr))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFound.Password), []byte(loginForm.Password)); err != nil {
		loginErr := &AuthError{
			Message: "Invalid username or password. Please try again.",
			Type:    "Authentication",
		}
		c.HTML(http.StatusForbidden, "", Login(loginErr))
	}

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userFound.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	// TODO: Use a real secret from env variables for signing jwt
	token, err := generateToken.SignedString([]byte("SECRET"))

	if err != nil {
		loginErr := &AuthError{
			Message: "There was a technical error. Please try again.",
			Type:    "Technical",
		}
		c.HTML(http.StatusBadRequest, "", Login(loginErr))
	}

	cookieName := "access_token"
	cookieMaxAge := 15 * 60
	secure := false
	sameSite := http.SameSiteLaxMode

	c.SetCookie(cookieName,
		token,
		cookieMaxAge,
		"/",
		"",
		secure,
		true)

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
	http.SetCookie(c.Writer, cookie)

	c.Redirect(http.StatusSeeOther, "/")
}

func (handler *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.Header("HX-Redirect", "/login")
}

func (handler *AuthHandler) CheckAccessToken(c *gin.Context) {
	ctx := context.Background()

	authToken, err := c.Cookie("access_token")

	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	tokenString := authToken
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// TODO: Use a real secret from env variables for signing jwt
		return []byte("SECRET"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, _ := handler.UserRepository.GetUserById(ctx, int64(claims["id"].(float64)))
	if user.ID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(
		"currentUser",
		user,
	)

	c.Next()
}

func (handler *AuthHandler) ResetPassword(c *gin.Context) {
	ctx := context.Background()

	var resetPassword ResetPassword
	c.Bind(&resetPassword)

	userFound, _ := handler.UserRepository.GetUserByEmailForLogin(ctx, resetPassword.Email)

	if userFound.ID == 0 {
		authErr := &AuthError{
			Message: "Invalid username. Please try again.",
			Type:    "Authentication",
		}
		c.HTML(http.StatusForbidden, "", ForgottenPassword(authErr))
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(resetPassword.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resetPasswordParams := entities.ResetPasswordParams{
		Email:    resetPassword.Email,
		Password: string(passwordHash),
	}

	resetErr := handler.UserRepository.ResetPassword(ctx, resetPasswordParams)

	if resetErr != nil {
		authErr := &AuthError{
			Message: "Unable to reset password. Please try again.",
			Type:    "Technical",
		}
		c.HTML(http.StatusForbidden, "", ForgottenPassword(authErr))
	}

	c.HTML(http.StatusOK, "", ResetPasswordSuccess())
}
