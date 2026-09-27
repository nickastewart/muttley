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
	"muttley/headtohead"
	"muttley/location"
	"muttley/mailer"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/user"

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
	var userRepository user.UserRepository = user.NewUserRepository(db)
	var locationRepository location.LocationRepository = location.NewLocationRepository(queries)
	var eventRepository event.EventRepository = event.NewEventRepository(queries)
	var eventResultRepository eventresult.EventResultRepository = eventresult.NewEventResultRepository(queries)
	var friendRepository friend.FriendRepository = friend.NewFriendRepository(queries)
	var dashboardRepository dashboard.DashboardRepository = dashboard.NewDashboardRepository(queries)
	var headToHeadRepository headtohead.HeadToHeadRepository = headtohead.NewHeadToHeadRepository(queries)
	magicLinks := auth.NewMagicLinkRepository(db)

	mail, err := mailer.NewFromEnv()
	if err != nil {
		log.Panic(err)
	}

	authHandler := auth.NewAuthHandler(userRepository, magicLinks, mail, sqlite.NewTransactor(db))
	userHandler := user.NewUserHandler(userRepository)
	fileUploadHandler := fileupload.NewFileUploadHandler(userRepository, eventRepository, locationRepository, eventResultRepository, sqlite.NewTransactor(db))
	eventHandler := event.NewEventsHandler(userRepository, eventRepository, locationRepository, eventResultRepository, friendRepository)
	friendHandler := friend.NewFriendHandler(friendRepository, userRepository)
	dashboardHandler := dashboard.NewDashboardHander(dashboardRepository, eventRepository, headToHeadRepository)
	headToHeadHandler := headtohead.NewHeadToHeadHandler(headToHeadRepository)

	router := gin.Default()
	router.Static("/styles", "./static/styles")
	router.Static("/images", "./static/images")
	router.Static("/icons", "./static/icons")

	router.POST("/signup", authHandler.RequestSignup)
	router.POST("/login", authHandler.RequestLogin)
	router.GET("/login/verify", authHandler.ShowVerify)
	router.POST("/login/verify", authHandler.VerifyMagicLink)
	router.POST("/logout", authHandler.CheckAccessToken, authHandler.Logout)
	auth.RegisterTestMagicLink(router, magicLinks)

	router.HTMLRender = &TemplRender{}

	router.GET("/", authHandler.CheckAccessToken, dashboardHandler.GetDashboard)
	router.GET("/dashboard", authHandler.CheckAccessToken, dashboardHandler.GetDashboard)

	router.GET("/login", authHandler.ShowLogin)
	router.GET("/signup", authHandler.ShowSignup)

	router.GET("/leaderboard", authHandler.CheckAccessToken, eventHandler.Leaderboard)

	router.GET("/upload", authHandler.CheckAccessToken, fileUploadHandler.UploadFile)
	router.POST("/upload/process", authHandler.CheckAccessToken, fileUploadHandler.ProcessFile)

	router.GET("/friends", authHandler.CheckAccessToken, friendHandler.Friends)
	router.GET("/head-to-head", authHandler.CheckAccessToken, headToHeadHandler.HeadToHead)
	router.GET("/search/friends", authHandler.CheckAccessToken, friendHandler.SearchFriends)

	router.POST("/remove-friend", authHandler.CheckAccessToken, friendHandler.RemoveFriend)
	router.POST("/add-friend", authHandler.CheckAccessToken, friendHandler.AddFriend)
	router.POST("/accept-friend", authHandler.CheckAccessToken, friendHandler.AcceptFriend)

	router.GET("/account", authHandler.CheckAccessToken, userHandler.Account)
	router.POST("/account", authHandler.CheckAccessToken, userHandler.UpdateAccount)
	router.POST("/account/delete", authHandler.CheckAccessToken, userHandler.DeleteAccount)

	router.Run()
}
