package user

import (
	"context"
	"database/sql"
	"muttley/sqlite"
	"muttley/sqlite/entities"
)

type UserRepositorySqlite struct {
	queries    *entities.Queries
	transactor *sqlite.Transactor
}

func NewUserRepository(db *sql.DB) *UserRepositorySqlite {
	return &UserRepositorySqlite{
		queries:    entities.New(db),
		transactor: sqlite.NewTransactor(db),
	}
}

func (r *UserRepositorySqlite) q(ctx context.Context) *entities.Queries {
	return sqlite.Queries(ctx, r.queries)
}

func (r *UserRepositorySqlite) GetUserById(ctx context.Context, id int64) (entities.GetUserByIdRow, error) {
	user, err := r.q(ctx).GetUserById(ctx, id)
	return user, err
}

func (r *UserRepositorySqlite) CreateUser(ctx context.Context,
	userParams entities.CreateUserParams) (entities.CreateUserRow, error) {
	savedUser, err := r.q(ctx).CreateUser(ctx, userParams)
	return savedUser, err
}

func (r *UserRepositorySqlite) GetUserByEmail(ctx context.Context, email string) (entities.GetUserByEmailRow, error) {
	user, err := r.q(ctx).GetUserByEmail(ctx, email)
	return user, err
}

func (r *UserRepositorySqlite) GetUserByEmailForLogin(ctx context.Context, email string) (entities.GetUserByEmailForLoginRow, error) {
	user, err := r.q(ctx).GetUserByEmailForLogin(ctx, email)
	return user, err
}

func (r *UserRepositorySqlite) GetUsersBySearchTerm(ctx context.Context, params entities.GetUsersBySearchTermParams) ([]entities.GetUsersBySearchTermRow, error) {
	users, err := r.q(ctx).GetUsersBySearchTerm(ctx, params)
	return users, err
}

func (r *UserRepositorySqlite) GetUserIdByProfileId(ctx context.Context, profileId string) (int64, error) {
	userId, err := r.q(ctx).GetUserIdByProfileId(ctx, profileId)
	return userId, err
}

func (r *UserRepositorySqlite) ResetPassword(ctx context.Context, params entities.ResetPasswordParams) error {
	err := r.q(ctx).ResetPassword(ctx, params)
	return err
}

func (r *UserRepositorySqlite) UpdateUser(ctx context.Context, params entities.UpdateUserParams) error {
	return r.q(ctx).UpdateUser(ctx, params)
}

// DeleteUser removes the user's results, friendships, and account in one
// transaction. A failure on any step puts the earlier deletes back.
func (r *UserRepositorySqlite) DeleteUser(ctx context.Context, id int64) error {
	return r.transactor.Within(ctx, func(ctx context.Context) error {
		q := r.q(ctx)
		if err := q.DeleteEventResultsByUserId(ctx, id); err != nil {
			return err
		}
		if err := q.DeleteFriendsByUserId(ctx, id); err != nil {
			return err
		}
		return q.DeleteUser(ctx, id)
	})
}
