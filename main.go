package main

import (
	"database/sql"
	_ "embed"
	"log"
	"muttley/auth"
	"muttley/dashboard"
	"muttley/event"
	"muttley/eventresult"
	"muttley/fileupload"
	"muttley/friend"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/templates"
	"muttley/user"
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
	var userRepository user.UserRepository = user.NewUserRepository(queries)
	var locationRepository location.LocationRepository = location.NewLocationRepository(queries)
	var eventRepository event.EventRepository = event.NewEventRepository(queries)
	var eventResultRepository eventresult.EventResultRepository = eventresult.NewEventResultRepository(queries)
	var friendRepository friend.FriendRepository = friend.NewFriendRepository(queries)
	var dashboardRepository dashboard.DashboardRepository = dashboard.NewDashboardRepository(queries)

	authHandler := auth.NewAuthHandler(userRepository)
	fileUploadHandler := fileupload.NewFileUploadHandler(userRepository, eventRepository, locationRepository, eventResultRepository)
	eventHandler := event.NewEventsHandler(userRepository, eventRepository, locationRepository, eventResultRepository, friendRepository)
	friendHandler := friend.NewFriendHandler(friendRepository, userRepository)
	dashboardHandler := dashboard.NewDashboardHander(dashboardRepository, eventRepository)

	router := gin.Default()
	router.Static("/styles", "./static/styles")
	router.Static("/images", "./static/images")
	router.Static("/icons", "./static/icons")

	router.POST("/signup", authHandler.Signup)
	router.POST("/login", authHandler.LoginForm)
	router.POST("/logout", authHandler.CheckAccessToken, authHandler.Logout)

	router.HTMLRender = &TemplRender{}

	router.GET("/", authHandler.CheckAccessToken, dashboardHandler.GetDashboard)
	router.GET("/dashboard", authHandler.CheckAccessToken, dashboardHandler.GetDashboard)

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
