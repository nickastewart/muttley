package friend

import (
	"context"
	"log/slog"
	"muttley/sqlite/entities"
)

type FriendRepositorySqlite struct {
	queries *entities.Queries
}

func NewFriendRepository(queries *entities.Queries) *FriendRepositorySqlite {
	return &FriendRepositorySqlite{
		queries: queries,
	}
}

func (r *FriendRepositorySqlite) AddFriend(ctx context.Context, addFriendParams entities.AddFriendParams) (entities.Friend, error) {
	friend, err := r.queries.AddFriend(ctx, addFriendParams)
	if err != nil {
		slog.Error(err.Error())
		return friend, err
	}
	return friend, nil
}

func (r *FriendRepositorySqlite) GetFriendsByUser(ctx context.Context, userId int64) ([]entities.GetFriendsByUserRow, error) {
	friends, err := r.queries.GetFriendsByUser(ctx, userId)
	if err != nil {
		slog.Error(err.Error())
	}
	return friends, err
}

func (r *FriendRepositorySqlite) UpdateFriendStatus(ctx context.Context, params entities.UpdateFriendStatusParams) (entities.Friend, error) {
	friend, err := r.queries.UpdateFriendStatus(ctx, params)
	if err != nil {
		slog.Error(err.Error())
	}
	return friend, err
}

func (r *FriendRepositorySqlite) GetFriendByUserIdAndFriendId(ctx context.Context, params entities.GetFriendByUserIdAndFriendIdParams) (entities.Friend, error) {
	friend, err := r.queries.GetFriendByUserIdAndFriendId(ctx, params)
	if err != nil {
		slog.Error(err.Error())
	}
	return friend, err
}
