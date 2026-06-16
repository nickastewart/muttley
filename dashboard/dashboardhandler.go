package dashboard

import (
	"context"
	"log/slog"
	"muttley/event"
	"muttley/sqlite/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	DashboardRepository DashboardRepository
	EventRepository     event.EventRepository
}

func NewDashboardHander(dashboardRepository DashboardRepository, eventRepository event.EventRepository) *DashboardHandler {
	return &DashboardHandler{
		DashboardRepository: dashboardRepository,
		EventRepository:     eventRepository,
	}
}

func (handler *DashboardHandler) GetDashboard(c *gin.Context) {
	ctx := context.Background()
	u, exists := c.Get("currentUser")

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not authenticated"})
		return
	}

	user := u.(entities.GetUserByIdRow)
	dashboard, err := handler.DashboardRepository.GetDashboard(ctx, user.ID)

	if err != nil {
		slog.Error("Error getting dashbaord")
		// TODO handle erorr in UI
	}

	bestTrack, err := handler.DashboardRepository.GetBestTrack(ctx, user.ID)
	if err != nil {
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

	c.Header("HX-Redirect", "/dashboard")
	c.HTML(http.StatusOK, "", Dashboard(dashboard, bestTrack.Name, locationStats, recentPositions, recentEvents))
}
