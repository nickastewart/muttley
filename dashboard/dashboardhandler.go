package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"muttley/event"
	"muttley/headtohead"
	"muttley/sqlite/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	DashboardRepository  DashboardRepository
	EventRepository      event.EventRepository
	HeadToHeadRepository headtohead.HeadToHeadRepository
}

func NewDashboardHander(dashboardRepository DashboardRepository, eventRepository event.EventRepository, headToHeadRepository headtohead.HeadToHeadRepository) *DashboardHandler {
	return &DashboardHandler{
		DashboardRepository:  dashboardRepository,
		EventRepository:      eventRepository,
		HeadToHeadRepository: headToHeadRepository,
	}
}

func (handler *DashboardHandler) GetDashboard(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.User)
	dashboard, err := handler.DashboardRepository.GetDashboard(ctx, user.ID)

	if err != nil {
		slog.Error("Error getting dashbaord")
		// TODO handle erorr in UI
	}

	bestTrack, err := handler.DashboardRepository.GetBestTrack(ctx, user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("Error getting bestTrack")
		// TODO handle erorr in UI
	}

	locationStats, err := handler.DashboardRepository.GetLocationStats(ctx, user.ID)
	if err != nil {
		slog.Error("Error getting bestTrack")
		// TODO handle erorr in UI
	}

	recentPositions, err := handler.DashboardRepository.GetRecentPositions(ctx, user.ID)
	if err != nil {
		slog.Error("Error getting bestTrack")
		// TODO handle erorr in UI
	}

	recentEvents, err := handler.EventRepository.GetRecentEvents(ctx, user.ID)
	if err != nil {
		slog.Error("Error getting recent events")
		// TODO handle erorr in UI
	}

	headToHeadWins := int64(0)
	headToHead, err := handler.HeadToHeadRepository.GetHeadToHead(ctx, user.ID)
	if err != nil {
		slog.Error("Error getting head to head wins")
		// TODO handle erorr in UI
	} else {
		for _, row := range headToHead {
			headToHeadWins += row.UserWins
		}
	}

	c.Header("HX-Redirect", "/dashboard")
	c.HTML(http.StatusOK, "", Dashboard(dashboard, bestTrack.Name, locationStats, recentPositions, recentEvents, headToHeadWins))
}
