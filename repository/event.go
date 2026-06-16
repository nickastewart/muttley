package repository

import (
	"context"
	"log/slog"
	"muttley/sqlite/entities"
)

type EventRepositorySqlite struct {
	queries *entities.Queries
}

func NewEventRepository(queries *entities.Queries) *EventRepositorySqlite {
	return &EventRepositorySqlite{
		queries: queries,
	}
}

func (r *EventRepositorySqlite) CreateEvent(ctx context.Context, arg entities.CreateEventParams) (entities.Event, error) {
	event, err := r.queries.CreateEvent(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
	}

	return event, err
}

func (r *EventRepositorySqlite) GetEventByLocationAndTypeAndDate(ctx context.Context, arg entities.GetEventByLocationAndTypeAndDateParams) (entities.Event, error) {
	event, err := r.queries.GetEventByLocationAndTypeAndDate(ctx, arg)

	if err != nil {
		slog.Error(err.Error())
	}

	return event, err
}

func (r *EventRepositorySqlite) GetEventsByUser(ctx context.Context, userID []int64) ([]entities.GetEventsByUserRow, error) {
	events, err := r.queries.GetEventsByUser(ctx, userID)
	if err != nil {
		slog.Error(err.Error())
	}
	return events, err
}

func (r *EventRepositorySqlite) GetRecentEvents(ctx context.Context, userId int64) ([]entities.GetRecentEventsRow, error) {
	events, err := r.queries.GetRecentEvents(ctx, userId)
	if err != nil {
		slog.Error(err.Error())
	}
	return events, err
}
