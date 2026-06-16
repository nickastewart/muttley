package event

import (
	"context"
	"muttley/sqlite/entities"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, arg entities.CreateEventParams) (entities.Event, error)
	GetEventByLocationAndTypeAndDate(ctx context.Context, arg entities.GetEventByLocationAndTypeAndDateParams) (entities.Event, error)
	GetEventsByUser(ctx context.Context, userIds []int64) ([]entities.GetEventsByUserRow, error)
	GetRecentEvents(ctx context.Context, userId int64) ([]entities.GetRecentEventsRow, error)
}
