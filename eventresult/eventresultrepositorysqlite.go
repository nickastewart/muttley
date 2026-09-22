package eventresult

import (
	"context"
	"log/slog"
	"muttley/sqlite"
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

func (r *EventResultRepositorySqlite) q(ctx context.Context) *entities.Queries {
	return sqlite.Queries(ctx, r.queries)
}

func (r *EventResultRepositorySqlite) CreateEventResult(ctx context.Context, arg entities.CreateEventResultParams) (entities.EventResult, error) {
	savedEntity, err := r.q(ctx).CreateEventResult(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
	}
	return savedEntity, err
}

func (r *EventResultRepositorySqlite) GetEventResultByEventIdAndUserId(ctx context.Context, arg entities.GetEventResultByEventIdAndUserIdParams) (entities.EventResult, error) {
	eventResult, err := r.q(ctx).GetEventResultByEventIdAndUserId(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
	}
	return eventResult, err
}

func (r *EventResultRepositorySqlite) GetUserFriendsResults(ctx context.Context, userId int64) ([]entities.GetUserFriendsResultsRow, error) {
	results, err := r.q(ctx).GetUserFriendsResults(ctx, userId)
	if err != nil {
		slog.Error(err.Error())
	}
	return results, err
}
