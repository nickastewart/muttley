package repository

import (
	"context"
	"muttley/sqlite/entities"
)

type UserRepositorySqlite struct {
	queries *entities.Queries
}

func NewUserRepository(queries *entities.Queries) *UserRepositorySqlite {
	return &UserRepositorySqlite{
		queries: queries,
	}
}

func (r *UserRepositorySqlite) GetUserById(ctx context.Context, id int64) (entities.GetUserByIdRow, error) {
	user, err := r.queries.GetUserById(ctx, id)
	return user, err
}

func (r *UserRepositorySqlite) CreateUser(ctx context.Context,
	userParams entities.CreateUserParams) (entities.CreateUserRow, error) {
	savedUser, err := r.queries.CreateUser(ctx, userParams)
	return savedUser, err
}

func (r *UserRepositorySqlite) GetUserByEmail(ctx context.Context, email string) (entities.GetUserByEmailRow, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	return user, err
}

func (r *UserRepositorySqlite) GetUserByEmailForLogin(ctx context.Context, email string) (entities.GetUserByEmailForLoginRow, error) {
	user, err := r.queries.GetUserByEmailForLogin(ctx, email)
	return user, err
}

func (r *UserRepositorySqlite) GetUsersBySearchTerm(ctx context.Context, params entities.GetUsersBySearchTermParams) ([]entities.GetUsersBySearchTermRow, error) {
	users, err := r.queries.GetUsersBySearchTerm(ctx, params)
	return users, err
}

func (r *UserRepositorySqlite) GetUserIdByProfileId(ctx context.Context, profileId string) (int64, error) {
	userId, err := r.queries.GetUserIdByProfileId(ctx, profileId)
	return userId, err
}

func (r *UserRepositorySqlite) ResetPassword(ctx context.Context, params entities.ResetPasswordParams) error {
	err := r.queries.ResetPassword(ctx, params)
	return err
}
