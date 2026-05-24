package main

import (
	"database/sql"
	_ "embed"
	"log"
	"muttley/auth"
	"muttley/dashboard"
	"muttley/handlers"
	"muttley/repository"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./sqlite/racer.db")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	queries := entities.New(db)
	var userRepository repository.UserRepository = repository.NewUserRepository(queries)
	var locationRepository repository.LocationRepository = repository.NewLocationRepository(queries)
	var eventRepository repository.EventRepository = repository.NewEventRepository(queries)
	var eventResultRepository repository.EventResultRepository = repository.NewEventResultRepository(queries)
	var friendRepository repository.FriendRepository = repository.NewFriendRepository(queries)

	authHandler := auth.NewAuthHandler(userRepository)
	fileUploadHandler := handlers.NewFileUploadHandler(userRepository, eventRepository, locationRepository, eventResultRepository)
	eventHandler := handlers.NewEventsHandler(userRepository, eventRepository, locationRepository, eventResultRepository, friendRepository)
	friendHandler := handlers.NewFriendHandler(friendRepository, userRepository)

	router := gin.Default()
	router.Static("/styles", "./static/styles")
	router.Static("/images", "./static/images")
	router.Static("/icons", "./static/icons")

	router.POST("/signup", authHandler.Signup)
	router.POST("/login", authHandler.LoginForm)
	router.POST("/logout", authHandler.CheckAccessToken, authHandler.Logout)

	router.HTMLRender = &TemplRender{}

	router.GET("/", authHandler.CheckAccessToken, func(c *gin.Context) {
		c.HTML(http.StatusOK, "", dashboard.Dashboard())
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", auth.Login(nil))
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Signup())
	})

	router.GET("/forgotten-password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", auth.ForgottenPassword(nil))
	})

	router.POST("/reset-password", authHandler.ResetPassword)

	router.GET("/leaderboard", authHandler.CheckAccessToken, eventHandler.Leaderboard)

	router.GET("/upload", authHandler.CheckAccessToken, fileUploadHandler.UploadFile)
	router.POST("/upload/process", authHandler.CheckAccessToken, fileUploadHandler.ProcessFile)

	router.GET("/friends", authHandler.CheckAccessToken, friendHandler.Friends)
	router.GET("/search/friends", authHandler.CheckAccessToken, friendHandler.SearchFriends)

	router.POST("/remove-friend", authHandler.CheckAccessToken, friendHandler.RemoveFriend)
	router.POST("/add-friend", authHandler.CheckAccessToken, friendHandler.AddFriend)

	router.Run()
}
