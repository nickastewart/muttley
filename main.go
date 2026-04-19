package main

import (
	"database/sql"
	_ "embed"
	"log"
	"muttley/controllers"
	"muttley/repository"
	"muttley/sqlite/entities"
	"muttley/templates"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func main() {
	// TODO: Add testing to parser

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

	authController := controllers.NewAuthController(userRepository)
	fileUploadController := controllers.NewFileUploadController(userRepository, eventRepository, locationRepository, eventResultRepository)
	eventController := controllers.NewEventsController(userRepository, eventRepository, locationRepository, eventResultRepository, friendRepository)
	friendHandler := controllers.NewFriendHandler(friendRepository, userRepository)

	if err != nil {
		log.Panic(err)
	}

	router := gin.Default()
	router.Static("/styles", "./static/styles")
	router.Static("/images", "./static/images")

	router.POST("/signup", authController.Signup)
	router.POST("/login", authController.LoginForm)

	router.HTMLRender = &TemplRender{}

	router.GET("/", authController.CheckAccessToken, func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Home())
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Login())
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Signup())
	})

	router.GET("/leaderboard", authController.CheckAccessToken, eventController.Leaderboard)

	router.GET("/upload", authController.CheckAccessToken, fileUploadController.UploadFile)
	router.POST("/upload/process", authController.CheckAccessToken, fileUploadController.ProcessFile)

	router.GET("/friends", authController.CheckAccessToken, friendHandler.Friends)
	// TODO: Create Search Users Page
	router.GET("/search/friends", authController.CheckAccessToken, friendHandler.SearchFriends)
	// TODO: Add endpoint to remove friends
	// TODO: Templating for adding friend (modal on top of the friends page)

	router.POST("/addFriend", authController.CheckAccessToken, friendHandler.AddFriend)

	router.Run()
}
