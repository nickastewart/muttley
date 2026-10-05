package fileupload

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"muttley/event"
	"muttley/eventresult"
	"muttley/location"
	"muttley/records"
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

type uploadRecords struct {
	PersonalBest bool
	TrackRecord  bool
}

var errBadResultsFile = errors.New("bad results file")

type FileUploadHandler struct {
	UserRepository         user.UserRepository
	LocationRepository     location.LocationRepository
	EventRepository        event.EventRepository
	EventResultRespository eventresult.EventResultRepository
	RecordsRepository      records.RecordsRepository
	Transactor             *sqlite.Transactor
}

func NewFileUploadHandler(userRepository user.UserRepository,
	eventRepository event.EventRepository,
	locationRepository location.LocationRepository,
	eventResultRespository eventresult.EventResultRepository,
	recordsRepository records.RecordsRepository,
	transactor *sqlite.Transactor) *FileUploadHandler {
	return &FileUploadHandler{
		UserRepository:         userRepository,
		EventRepository:        eventRepository,
		LocationRepository:     locationRepository,
		EventResultRespository: eventResultRespository,
		RecordsRepository:      recordsRepository,
		Transactor:             transactor,
	}
}

func (handler *FileUploadHandler) UploadFile(c *gin.Context) {
	c.Header("HX-Redirect", "/upload")
	renderUploadFile(c, "")
}

func (handler *FileUploadHandler) ProcessFile(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.User)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	form, err := c.MultipartForm()
	if err != nil {
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			renderUploadFile(c, "That file is too large.")
			return
		}
		renderUploadFile(c, "Choose one results file.")
		return
	}

	files := form.File["file"]
	if len(files) != 1 || files[0].Size == 0 {
		renderUploadFile(c, "Choose one results file.")
		return
	}

	file, err := files[0].Open()
	if err != nil {
		slog.Error("open results file", "error", err)
		renderUploadFile(c, "Unable to save this result. Please try again.")
		return
	}
	defer file.Close()

	parsed, err := parser.ParseFile(file)
	if err != nil {
		renderUploadFile(c, "This file does not look like a results email.")
		return
	}

	locationEntity, eventEntity, eventResultEntity, notes, err := handler.saveEvent(ctx, user, parsed)
	if err != nil {
		if errors.Is(err, errBadResultsFile) {
			renderUploadFile(c, "This file does not look like a results email.")
			return
		}
		slog.Error("save result", "error", err)
		renderUploadFile(c, "Unable to save this result. Please try again.")
		return
	}

	c.HTML(http.StatusOK, "", templates.UploadSuccess(locationEntity, eventEntity, eventResultEntity, notes.PersonalBest, notes.TrackRecord))
}

func renderUploadFile(c *gin.Context, errMsg string) {
	c.HTML(http.StatusOK, "", templates.UploadFile(errMsg))
}

func (handler *FileUploadHandler) saveEvent(ctx context.Context, currentUser entities.User, parsed *model.Event) (entities.Location, entities.Event, entities.EventResult, uploadRecords, error) {
	var savedLocation entities.Location
	var savedEvent entities.Event
	var saved entities.EventResult
	var notes uploadRecords
	if parsed.DriverInfo.Position < 1 || parsed.DriverInfo.Position > len(parsed.DriverTimes) {
		return savedLocation, savedEvent, saved, notes, errBadResultsFile
	}
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

		snapshot, err := handler.RecordsRepository.GetLocationRecordSnapshot(ctx, entities.GetLocationRecordSnapshotParams{
			Userid:     currentUser.ID,
			Locationid: locationEntity.ID,
		})
		if err != nil {
			log.Println("Failed to read records " + err.Error())
			return err
		}

		driverTime := parsed.DriverTimes[parsed.DriverInfo.Position-1]
		savedResult, created, err := handler.processEventResult(ctx, currentUser, &driverTime, &eventEntity)
		if err != nil {
			log.Println("Failed to process event result " + err.Error())
			return err
		}
		savedLocation = locationEntity
		savedEvent = eventEntity
		saved = savedResult
		notes = recordNotes(created, savedResult.BestLapTime, snapshot.PersonalBestLap, snapshot.TrackRecordLap)
		return nil
	})
	return savedLocation, savedEvent, saved, notes, err
}

func recordNotes(created bool, bestLap int64, previousPersonalBest int64, previousTrackRecord int64) uploadRecords {
	if !created || bestLap <= 0 {
		return uploadRecords{}
	}
	return uploadRecords{
		PersonalBest: previousPersonalBest == 0 || bestLap < previousPersonalBest,
		TrackRecord:  previousTrackRecord == 0 || bestLap < previousTrackRecord,
	}
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
	totalDrivers := int64(len(event.DriverTimes))
	candidates, err := handler.EventRepository.ListEventsByLocationAndTypeAndDate(ctx, entities.ListEventsByLocationAndTypeAndDateParams{
		LocationID: location.ID,
		Type:       event.RaceType,
		Date:       event.Date,
	})
	if err != nil {
		return entities.Event{}, err
	}

	if matched, ok := matchSession(candidates, totalDrivers); ok {
		return matched, nil
	}

	savedEvent, err := handler.EventRepository.CreateEvent(ctx, entities.CreateEventParams{
		LocationID:   location.ID,
		Type:         event.RaceType,
		Date:         event.Date,
		TotalDrivers: totalDrivers,
	})
	if err != nil {
		return savedEvent, err
	}
	return savedEvent, nil
}

func matchSession(candidates []entities.Event, totalDrivers int64) (entities.Event, bool) {
	if totalDrivers > 0 {
		for _, candidate := range candidates {
			if candidate.TotalDrivers == totalDrivers {
				return candidate, true
			}
		}
		return entities.Event{}, false
	}
	if len(candidates) == 0 {
		return entities.Event{}, false
	}
	return candidates[0], true
}

func (handler *FileUploadHandler) processEventResult(ctx context.Context, user entities.User, driverResult *model.DriverTime, event *entities.Event) (entities.EventResult, bool, error) {

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
			return eventResultEntity, false, err
		}
		return savedEntity, true, nil

	}
	return eventResultEntity, false, err
}
