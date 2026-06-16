package dashboard

import (
	"context"
	"log"
	"muttley/sqlite/entities"
)

type DashboardRepositorySqlite struct {
	queries *entities.Queries
}

func NewDashboardRepository(queries *entities.Queries) *DashboardRepositorySqlite {
	return &DashboardRepositorySqlite{
		queries: queries,
	}
}

func (r *DashboardRepositorySqlite) GetDashboard(context context.Context, userId int64) (entities.GetDashboardRow, error) {
	dashboard, err := r.queries.GetDashboard(context, userId)
	if err != nil {
		log.Println(err)
	}
	return dashboard, err
}

func (r *DashboardRepositorySqlite) GetBestTrack(context context.Context, userId int64) (entities.GetBestTrackRow, error) {
	bestTrack, err := r.queries.GetBestTrack(context, userId)
	if err != nil {
		log.Println(err)
	}
	return bestTrack, err
}

func (r *DashboardRepositorySqlite) GetLocationStats(context context.Context, userId int64) ([]entities.GetLocationStatsRow, error) {
	locationStats, err := r.queries.GetLocationStats(context, userId)
	if err != nil {
		log.Println(err)
	}
	return locationStats, err
}

func (r *DashboardRepositorySqlite) GetRecentPositions(context context.Context, userId int64) ([]int64, error) {
	recentPositions, err := r.queries.GetRecentPositions(context, userId)
	if err != nil {
		log.Println(err)
	}
	return recentPositions, err
}
