package repository

import (
	"context"
	"log"
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
		log.Println(err)
	}

	return event, err
}

func (r *EventRepositorySqlite) GetEventByLocationAndTypeAndDate(ctx context.Context, arg entities.GetEventByLocationAndTypeAndDateParams) (entities.Event, error) {
	event, err := r.queries.GetEventByLocationAndTypeAndDate(ctx, arg)

	if err != nil {
		log.Println(err)
	}

	return event, err
}

func (r *EventRepositorySqlite) GetEventsByUser(ctx context.Context, userID []int64) ([]entities.GetEventsByUserRow, error) {
	events, err := r.queries.GetEventsByUser(ctx, userID)
	if err != nil {
		log.Println(err)
	}
	return events, err
}
