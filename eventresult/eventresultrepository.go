package eventresult

import (
	"context"
	"muttley/sqlite/entities"
)

type EventResultRepository interface {
	CreateEventResult(ctx context.Context, arg entities.CreateEventResultParams) (entities.EventResult, error)
	GetEventResultByEventIdAndUserId(ctx context.Context, arg entities.GetEventResultByEventIdAndUserIdParams) (entities.EventResult, error)
	GetUserFriendsResults(ctx context.Context, userId int64) ([]entities.GetUserFriendsResultsRow, error)
}
