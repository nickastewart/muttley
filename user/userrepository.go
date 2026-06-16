package user

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
