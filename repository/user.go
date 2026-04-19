package repository

import (
	"context"
	"log"
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

	if err != nil {
		log.Println(err)
	}
	return user, err
}

func (r *UserRepositorySqlite) CreateUser(ctx context.Context,
	userParams entities.CreateUserParams) (entities.CreateUserRow, error) {
	savedUser, err := r.queries.CreateUser(ctx, userParams)

	if err != nil {
		log.Println(err)
	}

	return savedUser, err
}

func (r *UserRepositorySqlite) GetUserByEmail(ctx context.Context, email string) (entities.GetUserByEmailRow, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)

	if err != nil {
		log.Println(err)
	}

	return user, err
}

func (r *UserRepositorySqlite) GetUserByEmailForLogin(ctx context.Context, email string) (entities.GetUserByEmailForLoginRow, error) {
	user, err := r.queries.GetUserByEmailForLogin(ctx, email)

	if err != nil {
		log.Println(err)
	}

	return user, err
}

func (r *UserRepositorySqlite) GetUsersBySearchTerm(ctx context.Context, params entities.GetUsersBySearchTermParams) ([]entities.GetUsersBySearchTermRow, error) {
	users, err := r.queries.GetUsersBySearchTerm(ctx, params)
	if err != nil {
		log.Panic(err)
	}
	return users, err
}
