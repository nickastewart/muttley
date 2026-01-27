package controllers

import (
	"database/sql"
	"errors"
	"log"
	"muttley/repository"
	"muttley/sqlite/entities"
	"muttley/templates"

	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	parser "github.com/nickastewart/muttley-parser"
	"github.com/nickastewart/muttley-parser/model"
)

type FileUploadController struct {
	UserRepository         repository.UserRepository
	LocationRepository     repository.LocationRepository
	EventRepository        repository.EventRepository
	EventResultRespository repository.EventResultRepository
}

func NewFileUploadController(userRepository repository.UserRepository,
	eventRepository repository.EventRepository,
	locationRepository repository.LocationRepository,
	eventResultRespository repository.EventResultRepository) *FileUploadController {
	return &FileUploadController{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
	}
}

func (controller *FileUploadController) UploadFile(c *gin.Context) {
	c.HTML(http.StatusOK, "", templates.UploadFile())
}

func (controller *FileUploadController) ProcessFile(c *gin.Context) {
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

	locationEntity, err := controller.processLocation(ctx, event)

	if err != nil {
		log.Println("Failed to process location " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventEntity, err := controller.processEvent(ctx, event, &locationEntity)
	if err != nil {
		log.Println("Failed to process event " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverTime := event.DriverTimes[event.DriverInfo.Position-1]
	eventResultEntity, err := controller.processEventResult(ctx, user, &driverTime, &eventEntity)

	if err != nil {
		log.Println("Failed to process event result " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": eventResultEntity})
}

func (controller *FileUploadController) processLocation(ctx context.Context, event *model.Event) (entities.Location, error) {

	location, err := controller.LocationRepository.GetLocationByName(ctx, event.Location)

	if location.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createdLocation, err := controller.LocationRepository.CreateLocation(ctx, event.Location)
		if err != nil {
			return location, err
		}
		location = createdLocation
	}

	return location, err
}

func (controller *FileUploadController) processEvent(ctx context.Context, event *model.Event, location *entities.Location) (entities.Event, error) {

	getEventParams := entities.GetEventByLocationAndTypeAndDateParams{
		LocationID: location.ID,
		Type:       event.RaceType,
		Date:       event.Date,
	}

	eventEntity, err := controller.EventRepository.GetEventByLocationAndTypeAndDate(ctx, getEventParams)

	if eventEntity.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createEventParams := entities.CreateEventParams{
			LocationID:   location.ID,
			Type:         event.RaceType,
			Date:         event.Date,
			TotalDrivers: int64(len(event.DriverTimes)),
		}

		savedEvent, err := controller.EventRepository.CreateEvent(ctx, createEventParams)
		if err != nil {
			return eventEntity, err
		}
		eventEntity = savedEvent
	}

	return eventEntity, err
}

func (controller *FileUploadController) processEventResult(ctx context.Context, user entities.GetUserByIdRow, driverResult *model.DriverTime, event *entities.Event) (entities.EventResult, error) {

	getEventResultByEventIdAndUserIdParams := entities.GetEventResultByEventIdAndUserIdParams{
		EventID: event.ID,
		UserID:  user.ID,
	}

	eventResultEntity, err := controller.EventResultRespository.GetEventResultByEventIdAndUserId(ctx, getEventResultByEventIdAndUserIdParams)

	if eventResultEntity.ID == 0 || errors.Is(err, sql.ErrNoRows) {
		createEventResultParams := entities.CreateEventResultParams{
			EventID:        event.ID,
			UserID:         user.ID,
			BestLapTime:    int64(driverResult.Best),
			AverageLapTime: int64(driverResult.Avg),
			Position:       int64(driverResult.Pos),
			NumberOfLaps:   int64(driverResult.NoLaps),
		}

		savedEntity, err := controller.EventResultRespository.CreateEventResult(ctx, createEventResultParams)

		if err != nil {
			return eventResultEntity, err
		}

		eventResultEntity = savedEntity
	}
	return eventResultEntity, err
}
