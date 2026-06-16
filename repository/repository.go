package repository

import (
	"context"
	"muttley/sqlite/entities"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id int64) (entities.GetUserByIdRow, error)
	GetUserByEmail(ctx context.Context, email string) (entities.GetUserByEmailRow, error)
	GetUserByEmailForLogin(ctx context.Context, email string) (entities.GetUserByEmailForLoginRow, error)
	CreateUser(ctx context.Context, createUserParams entities.CreateUserParams) (entities.CreateUserRow, error)
	GetUsersBySearchTerm(ctx context.Context, params entities.GetUsersBySearchTermParams) ([]entities.GetUsersBySearchTermRow, error)
	GetUserIdByProfileId(ctx context.Context, profileId string) (int64, error)
	ResetPassword(ctx context.Context, params entities.ResetPasswordParams) error
}

type LocationRepository interface {
	CreateLocation(ctx context.Context, name string) (entities.Location, error)
	GetLocationByName(ctx context.Context, name string) (entities.Location, error)
}

type EventRepository interface {
	CreateEvent(ctx context.Context, arg entities.CreateEventParams) (entities.Event, error)
	GetEventByLocationAndTypeAndDate(ctx context.Context, arg entities.GetEventByLocationAndTypeAndDateParams) (entities.Event, error)
	GetEventsByUser(ctx context.Context, userIds []int64) ([]entities.GetEventsByUserRow, error)
	GetRecentEvents(ctx context.Context, userId int64) ([]entities.GetRecentEventsRow, error)
}

type EventResultRepository interface {
	CreateEventResult(ctx context.Context, arg entities.CreateEventResultParams) (entities.EventResult, error)
	GetEventResultByEventIdAndUserId(ctx context.Context, arg entities.GetEventResultByEventIdAndUserIdParams) (entities.EventResult, error)
	GetUserFriendsResults(ctx context.Context, userId int64) ([]entities.GetUserFriendsResultsRow, error)
}

type FriendRepository interface {
	AddFriend(ctx context.Context, arg entities.AddFriendParams) (entities.Friend, error)
	GetFriendsByUser(ctx context.Context, userId int64) ([]entities.GetFriendsByUserRow, error)
	UpdateFriendStatus(ctx context.Context, params entities.UpdateFriendStatusParams) (entities.Friend, error)
	GetFriendByUserIdAndFriendId(ctx context.Context, args entities.GetFriendByUserIdAndFriendIdParams) (entities.Friend, error)
}
