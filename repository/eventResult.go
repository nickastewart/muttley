package repository

import (
	"context"
	"log"
	"muttley/sqlite/entities"
)

type EventResultRepositorySqlite struct {
	queries *entities.Queries
}

func NewEventResultRepository(queries *entities.Queries) *EventResultRepositorySqlite {
	return &EventResultRepositorySqlite{
		queries: queries,
	}
}

func (r *EventResultRepositorySqlite) CreateEventResult(ctx context.Context, arg entities.CreateEventResultParams) (entities.EventResult, error) {
	savedEntity, err := r.queries.CreateEventResult(ctx, arg)
	if err != nil {
		log.Println(err.Error())
	}
	return savedEntity, err
}

func (r *EventResultRepositorySqlite) GetEventResultByEventIdAndUserId(ctx context.Context, arg entities.GetEventResultByEventIdAndUserIdParams) (entities.EventResult, error) {
	eventResult, err := r.queries.GetEventResultByEventIdAndUserId(ctx, arg)
	if err != nil {
		log.Println(err.Error())
	}
	return eventResult, err
}
