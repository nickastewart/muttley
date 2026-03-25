package repository

import (
	"context"
	"log"
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
		log.Println(err)
	}
	return friend, nil
}

func (r *FriendRepositorySqlite) GetFriendsByUser(ctx context.Context, userId int64) ([]entities.GetFriendsByUserRow, error) {
	params := entities.GetFriendsByUserParams{
		UserID:   userId,
		FriendID: userId,
	}
	friends, err := r.queries.GetFriendsByUser(ctx, params)
	if err != nil {
		log.Println(err)
	}
	return friends, err
}
