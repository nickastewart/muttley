package dashboard

import (
	"context"
	"muttley/sqlite/entities"
)

type DashboardRepository interface {
	GetDashboard(ctx context.Context, userId int64) (entities.GetDashboardRow, error)
	GetBestTrack(ctx context.Context, userId int64) (entities.GetBestTrackRow, error)
	GetLocationStats(ctx context.Context, userId int64) ([]entities.GetLocationStatsRow, error)
	GetRecentPositions(ctx context.Context, userId int64) ([]int64, error)
}
