package fileupload

import (
	"database/sql"
	"errors"
	"log"
	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/sqlite"
	"muttley/sqlite/entities"
	"muttley/templates"
	"muttley/user"

	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	parser "github.com/nickastewart/muttley-parser"
	"github.com/nickastewart/muttley-parser/model"
)

type FileUploadHandler struct {
	UserRepository         user.UserRepository
	LocationRepository     location.LocationRepository
	EventRepository        event.EventRepository
	EventResultRespository eventresult.EventResultRepository
	Transactor             *sqlite.Transactor
}

func NewFileUploadHandler(userRepository user.UserRepository,
	eventRepository event.EventRepository,
	locationRepository location.LocationRepository,
	eventResultRespository eventresult.EventResultRepository,
	transactor *sqlite.Transactor) *FileUploadHandler {
	return &FileUploadHandler{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
		Transactor:             transactor,
	}
}

func (handler *FileUploadHandler) UploadFile(c *gin.Context) {
	c.Header("HX-Redirect", "/upload")
	c.HTML(http.StatusOK, "", templates.UploadFile())
}

func (handler *FileUploadHandler) ProcessFile(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)

	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	multipartFile := form.File["file"]
	log.Println(multipartFile)
	file, err := multipartFile[0].Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	event, err := parser.ParseFile(file)
	if err != nil {
		log.Println("Failed to parse file")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventResultEntity, err := handler.saveEvent(ctx, user, event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": eventResultEntity})
}

func (handler *FileUploadHandler) saveEvent(ctx context.Context, currentUser entities.GetUserByIdRow, parsed *model.Event) (entities.EventResult, error) {
	var saved entities.EventResult
	err := handler.Transactor.Within(ctx, func(ctx context.Context) error {
		locationEntity, err := handler.processLocation(ctx, parsed)
		if err != nil {
			log.Println("Failed to process location " + err.Error())
			return err
		}

		eventEntity, err := handler.processEvent(ctx, parsed, &locationEntity)
		if err != nil {
			log.Println("Failed to process event " + err.Error())
			return err
		}

		driverTime := parsed.DriverTimes[parsed.DriverInfo.Position-1]
		saved, err = handler.processEventResult(ctx, currentUser, &driverTime, &eventEntity)
		if err != nil {
			log.Println("Failed to process event result " + err.Error())
			return err
		}
		return nil
	})
	return saved, err
}

func (handler *FileUploadHandler) processLocation(ctx context.Context, event *model.Event) (entities.Location, error) {

	location, err := handler.LocationRepository.GetLocationByName(ctx, event.Location)

	if location.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createdLocation, err := handler.LocationRepository.CreateLocation(ctx, event.Location)
		if err != nil {
			return createdLocation, err
		}
		return createdLocation, nil
	}

	return location, err
}

func (handler *FileUploadHandler) processEvent(ctx context.Context, event *model.Event, location *entities.Location) (entities.Event, error) {

	getEventParams := entities.GetEventByLocationAndTypeAndDateParams{
		LocationID: location.ID,
		Type:       event.RaceType,
		Date:       event.Date,
	}

	eventEntity, err := handler.EventRepository.GetEventByLocationAndTypeAndDate(ctx, getEventParams)

	if eventEntity.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createEventParams := entities.CreateEventParams{
			LocationID:   location.ID,
			Type:         event.RaceType,
			Date:         event.Date,
			TotalDrivers: int64(len(event.DriverTimes)),
		}

		savedEvent, err := handler.EventRepository.CreateEvent(ctx, createEventParams)
		if err != nil {
			return savedEvent, err
		}
		return savedEvent, nil
	}

	return eventEntity, err
}

func (handler *FileUploadHandler) processEventResult(ctx context.Context, user entities.GetUserByIdRow, driverResult *model.DriverTime, event *entities.Event) (entities.EventResult, error) {

	getEventResultByEventIdAndUserIdParams := entities.GetEventResultByEventIdAndUserIdParams{
		EventID: event.ID,
		UserID:  user.ID,
	}

	eventResultEntity, err := handler.EventResultRespository.GetEventResultByEventIdAndUserId(ctx, getEventResultByEventIdAndUserIdParams)

	if eventResultEntity.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createEventResultParams := entities.CreateEventResultParams{
			EventID:        event.ID,
			UserID:         user.ID,
			BestLapTime:    int64(driverResult.Best),
			AverageLapTime: int64(driverResult.Avg),
			Position:       int64(driverResult.Pos),
			NumberOfLaps:   int64(driverResult.NoLaps),
		}

		savedEntity, err := handler.EventResultRespository.CreateEventResult(ctx, createEventResultParams)

		if err != nil {
			return eventResultEntity, err
		}
		return savedEntity, nil

	}
	return eventResultEntity, err
}
