package friend

import (
	"context"
	"muttley/sqlite/entities"
)

type FriendRepository interface {
	AddFriend(ctx context.Context, arg entities.AddFriendParams) (entities.Friend, error)
	GetFriendsByUser(ctx context.Context, userId int64) ([]entities.GetFriendsByUserRow, error)
	UpdateFriendStatus(ctx context.Context, params entities.UpdateFriendStatusParams) (entities.Friend, error)
	GetFriendByUserIdAndFriendId(ctx context.Context, args entities.GetFriendByUserIdAndFriendIdParams) (entities.Friend, error)
}
