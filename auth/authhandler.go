package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"muttley/mailer"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AuthHandler struct {
	UserRepository      user.UserRepository
	MagicLinkRepository MagicLinkRepository
	Mailer              mailer.Mailer
	Transactor          *sqlite.Transactor
}

func NewAuthHandler(userRepository user.UserRepository, magicLinks MagicLinkRepository, mailer mailer.Mailer, transactor *sqlite.Transactor) *AuthHandler {
	return &AuthHandler{
		UserRepository:      userRepository,
		MagicLinkRepository: magicLinks,
		Mailer:              mailer,
		Transactor:          transactor,
	}
}

func (handler *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "", Login(Page{}))
}

func (handler *AuthHandler) ShowSignup(c *gin.Context) {
	email, ok := normalizeEmail(c.Query("email"))
	if !ok {
		email = ""
	}
	c.HTML(http.StatusOK, "", Signup(Page{Email: email}))
}

func (handler *AuthHandler) RequestLogin(c *gin.Context) {
	ctx := c.Request.Context()
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError()}))
		return
	}

	email, ok := normalizeEmail(form.Email)
	if !ok {
		c.HTML(http.StatusBadRequest, "", Login(Page{
			Error: &AuthError{Type: "Authentication", Message: "Enter a valid email address."},
			Email: strings.TrimSpace(form.Email),
		}))
		return
	}

	found, err := handler.UserRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError(), Email: email}))
		return
	}
	if found.ID == 0 {
		c.HTML(http.StatusBadRequest, "", Login(Page{
			Error: &AuthError{Type: "Authentication", Message: "No account found for that email."},
			Email: email,
		}))
		return
	}

	if err := handler.issueMagicLink(ctx, c, email, purposeLogin); err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError(), Email: email}))
		return
	}
	c.HTML(http.StatusOK, "", Login(Page{SentEmail: email, Email: email}))
}

func (handler *AuthHandler) RequestSignup(c *gin.Context) {
	ctx := c.Request.Context()
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "", Signup(Page{Error: technicalError()}))
		return
	}

	email, ok := normalizeEmail(form.Email)
	if !ok {
		c.HTML(http.StatusBadRequest, "", Signup(Page{
			Error: &AuthError{Type: "Authentication", Message: "Enter a valid email address."},
			Email: strings.TrimSpace(form.Email),
		}))
		return
	}

	found, err := handler.UserRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.HTML(http.StatusBadRequest, "", Signup(Page{Error: technicalError(), Email: email}))
		return
	}
	if found.ID != 0 {
		c.HTML(http.StatusBadRequest, "", Signup(Page{
			Error: &AuthError{Type: "Authentication", Message: "An account with this email already exists."},
			Email: email,
		}))
		return
	}

	if err := handler.issueMagicLink(ctx, c, email, purposeSignup); err != nil {
		c.HTML(http.StatusBadRequest, "", Signup(Page{Error: technicalError(), Email: email}))
		return
	}
	c.HTML(http.StatusOK, "", Signup(Page{SentEmail: email, Email: email}))
}

func (handler *AuthHandler) ShowVerify(c *gin.Context) {
	link, err := handler.activeLink(c.Request.Context(), c.Query("token"))
	if errors.Is(err, errInvalidLink) {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: invalidLinkError()}))
		return
	}
	if err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError()}))
		return
	}
	c.HTML(http.StatusOK, "", VerifyLink(VerifyPage{
		Token:    c.Query("token"),
		IsSignup: link.Purpose == purposeSignup,
	}))
}

func (handler *AuthHandler) VerifyMagicLink(c *gin.Context) {
	ctx := c.Request.Context()
	var form struct {
		Token string `form:"token"`
	}
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError()}))
		return
	}

	var userID int64
	err := handler.Transactor.Within(ctx, func(ctx context.Context) error {
		link, err := handler.activeLink(ctx, form.Token)
		if err != nil {
			return err
		}
		consumed, err := handler.MagicLinkRepository.Consume(ctx, link.ID)
		if err != nil {
			return err
		}
		if consumed != 1 {
			return errInvalidLink
		}
		id, err := handler.userIDForLink(ctx, link)
		if err != nil {
			return err
		}
		userID = id
		return nil
	})
	if errors.Is(err, errInvalidLink) {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: invalidLinkError()}))
		return
	}
	if err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError()}))
		return
	}

	if err := handler.setSession(c, userID); err != nil {
		c.HTML(http.StatusBadRequest, "", Login(Page{Error: technicalError()}))
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func (handler *AuthHandler) issueMagicLink(ctx context.Context, c *gin.Context, email, purpose string) error {
	raw, hash, err := newMagicToken()
	if err != nil {
		return err
	}
	expires := time.Now().UTC().Add(magicLinkTTL).Format(time.RFC3339)
	err = handler.Transactor.Within(ctx, func(ctx context.Context) error {
		if err := handler.MagicLinkRepository.InvalidateUnused(ctx, email); err != nil {
			return err
		}
		_, err := handler.MagicLinkRepository.Create(ctx, entities.CreateMagicLinkParams{
			Email:     email,
			TokenHash: hash,
			Purpose:   purpose,
			ExpiresAt: expires,
		})
		return err
	})
	if err != nil {
		return err
	}
	return handler.Mailer.SendMagicLink(ctx, email, magicLinkURL(c, raw))
}

func (handler *AuthHandler) activeLink(ctx context.Context, token string) (entities.MagicLink, error) {
	if strings.TrimSpace(token) == "" {
		return entities.MagicLink{}, errInvalidLink
	}
	link, err := handler.MagicLinkRepository.GetActiveByTokenHash(ctx, entities.GetActiveMagicLinkByTokenHashParams{
		TokenHash: hashMagicToken(token),
		Now:       time.Now().UTC().Format(time.RFC3339),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return entities.MagicLink{}, errInvalidLink
	}
	if err != nil {
		return entities.MagicLink{}, err
	}
	return link, nil
}

func (handler *AuthHandler) userIDForLink(ctx context.Context, link entities.MagicLink) (int64, error) {
	found, err := handler.UserRepository.GetUserByEmail(ctx, link.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if found.ID != 0 {
		return found.ID, nil
	}
	if link.Purpose != purposeSignup {
		return 0, errInvalidLink
	}

	_, err = handler.UserRepository.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   "",
		LastName:    "",
		Email:       link.Email,
		ProfileID:   profileIDForEmail(link.Email),
		DisplayName: "",
	})
	if err != nil {
		return 0, err
	}
	created, err := handler.UserRepository.GetUserByEmail(ctx, link.Email)
	if err != nil {
		return 0, err
	}
	if created.ID == 0 {
		return 0, fmt.Errorf("created user %s has no id", link.Email)
	}
	return created.ID, nil
}

func (handler *AuthHandler) setSession(c *gin.Context, userID int64) error {
	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	// TODO: Use a real secret from env variables for signing jwt
	token, err := generateToken.SignedString([]byte("SECRET"))
	if err != nil {
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
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
