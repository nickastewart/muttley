package user_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/event"
	"muttley/eventresult"
	"muttley/friend"
	"muttley/location"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestCreateAndGetUser(t *testing.T) {
	repo := newUserRepo(t)
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   "Ada",
		LastName:    "Lovelace",
		Email:       "ada@example.com",
		Password:    "secret",
		ProfileID:   "ada",
		DisplayName: "Ada",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.FirstName != "Ada" || created.LastName != "Lovelace" || created.Email != "ada@example.com" || created.ProfileID != "ada" || created.DisplayName != "Ada" {
		t.Fatalf("created row = %+v", created)
	}
	if !created.CreatedAt.Valid || created.CreatedAt.String == "" {
		t.Fatalf("created_at = %+v, want a timestamp", created.CreatedAt)
	}

	byEmail, err := repo.GetUserByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if byEmail.FirstName != "Ada" || byEmail.ProfileID != "ada" {
		t.Fatalf("user by email = %+v", byEmail)
	}

	byID, err := repo.GetUserById(ctx, byEmail.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if byID.Email != "ada@example.com" || byID.DisplayName != "Ada" || !byID.CreatedAt.Valid {
		t.Fatalf("user by id = %+v", byID)
	}

	login, err := repo.GetUserByEmailForLogin(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get user for login: %v", err)
	}
	if login.ID != byEmail.ID || login.Email != "ada@example.com" || login.Password != "secret" {
		t.Fatalf("login row = %+v", login)
	}

	profileID, err := repo.GetUserIdByProfileId(ctx, "ada")
	if err != nil {
		t.Fatalf("get user id by profile: %v", err)
	}
	if profileID != byEmail.ID {
		t.Fatalf("profile id lookup = %d, want %d", profileID, byEmail.ID)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := newUserRepo(t)
	ctx := context.Background()
	existing := createUser(t, repo, "Ada", "Lovelace", "ada@example.com", "ada")

	err := repo.UpdateUser(ctx, entities.UpdateUserParams{
		FirstName:   "Augusta",
		LastName:    "King",
		Email:       "augusta@example.com",
		DisplayName: "Countess",
		ID:          existing.ID,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	updated, err := repo.GetUserById(ctx, existing.ID)
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if updated.FirstName != "Augusta" || updated.LastName != "King" || updated.Email != "augusta@example.com" || updated.DisplayName != "Countess" || updated.ProfileID != "ada" {
		t.Fatalf("updated user = %+v", updated)
	}

	login, err := repo.GetUserByEmailForLogin(ctx, "augusta@example.com")
	if err != nil {
		t.Fatalf("get login after update: %v", err)
	}
	if login.Password != "secret" {
		t.Fatalf("password = %q, want it unchanged", login.Password)
	}
}

func TestResetPassword(t *testing.T) {
	repo := newUserRepo(t)
	ctx := context.Background()
	createUser(t, repo, "Ada", "Lovelace", "ada@example.com", "ada")

	err := repo.ResetPassword(ctx, entities.ResetPasswordParams{
		Password: "new-secret",
		Email:    "ada@example.com",
	})
	if err != nil {
		t.Fatalf("reset password: %v", err)
	}

	login, err := repo.GetUserByEmailForLogin(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get login: %v", err)
	}
	if login.Password != "new-secret" {
		t.Fatalf("password = %q, want new-secret", login.Password)
	}
}

func TestSearchUsersByName(t *testing.T) {
	db := testdb.Open(t)
	repo := user.NewUserRepository(entities.New(db))
	friends := friend.NewFriendRepository(entities.New(db))
	ctx := context.Background()

	searcher := createUser(t, repo, "Nick", "Stewart", "nick@example.com", "nick")
	anna := createUser(t, repo, "Anna", "Stewart", "anna@example.com", "anna")
	createUser(t, repo, "Bob", "Jones", "bob@example.com", "bob")
	cara := createUser(t, repo, "Cara", "Stone", "cara@example.com", "cara")

	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       cara.ID,
		FriendID:     searcher.ID,
		FriendStatus: "REQUESTED",
	}); err != nil {
		t.Fatalf("add friend: %v", err)
	}

	matches, err := repo.GetUsersBySearchTerm(ctx, entities.GetUsersBySearchTermParams{
		Name:   "%Stewart%",
		Userid: searcher.ID,
	})
	if err != nil {
		t.Fatalf("search stewart: %v", err)
	}
	if len(matches) != 1 || matches[0].ID != anna.ID || matches[0].FriendStatus != "NONE" {
		t.Fatalf("stewart matches = %+v", matches)
	}

	caraMatches, err := repo.GetUsersBySearchTerm(ctx, entities.GetUsersBySearchTermParams{
		Name:   "%cara%",
		Userid: searcher.ID,
	})
	if err != nil {
		t.Fatalf("search cara: %v", err)
	}
	if len(caraMatches) != 1 || caraMatches[0].ID != cara.ID || caraMatches[0].FriendStatus != "REQUESTED" {
		t.Fatalf("cara matches = %+v", caraMatches)
	}

	none, err := repo.GetUsersBySearchTerm(ctx, entities.GetUsersBySearchTermParams{
		Name:   "%nobody%",
		Userid: searcher.ID,
	})
	if err != nil {
		t.Fatalf("search nobody: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("nobody matches = %+v, want none", none)
	}
}

func TestDeleteUserRemovesResultsAndFriendships(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	users := user.NewUserRepository(entities.New(db))
	locations := location.NewLocationRepository(entities.New(db))
	events := event.NewEventRepository(entities.New(db))
	results := eventresult.NewEventResultRepository(entities.New(db))
	friends := friend.NewFriendRepository(entities.New(db))

	owner := createUser(t, users, "Ada", "Lovelace", "ada@example.com", "ada")
	other := createUser(t, users, "Grace", "Hopper", "grace@example.com", "grace")
	kept := createUser(t, users, "Alan", "Turing", "alan@example.com", "alan")

	track, err := locations.CreateLocation(ctx, "Whilton Mill")
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	race, err := events.CreateEvent(ctx, entities.CreateEventParams{
		LocationID:   track.ID,
		Type:         "Rental",
		Date:         "2024-06-01",
		TotalDrivers: 8,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := results.CreateEventResult(ctx, entities.CreateEventResultParams{
		EventID:        race.ID,
		UserID:         owner.ID,
		BestLapTime:    45000,
		AverageLapTime: 47000,
		Position:       1,
		NumberOfLaps:   10,
	}); err != nil {
		t.Fatalf("create owner result: %v", err)
	}
	if _, err := results.CreateEventResult(ctx, entities.CreateEventResultParams{
		EventID:        race.ID,
		UserID:         other.ID,
		BestLapTime:    46000,
		AverageLapTime: 48000,
		Position:       2,
		NumberOfLaps:   10,
	}); err != nil {
		t.Fatalf("create other result: %v", err)
	}
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       owner.ID,
		FriendID:     other.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add owner friend: %v", err)
	}
	keptFriend, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       other.ID,
		FriendID:     kept.ID,
		FriendStatus: "ACCEPTED",
	})
	if err != nil {
		t.Fatalf("add kept friend: %v", err)
	}

	if err := users.DeleteUser(ctx, owner.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err = users.GetUserById(ctx, owner.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted user error = %v, want sql.ErrNoRows", err)
	}
	_, err = results.GetEventResultByEventIdAndUserId(ctx, entities.GetEventResultByEventIdAndUserIdParams{
		EventID: race.ID,
		UserID:  owner.ID,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted result error = %v, want sql.ErrNoRows", err)
	}
	_, err = friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   owner.ID,
		Friendid: other.ID,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted friendship error = %v, want sql.ErrNoRows", err)
	}

	if _, err := users.GetUserById(ctx, other.ID); err != nil {
		t.Fatalf("other user was removed: %v", err)
	}
	if _, err := results.GetEventResultByEventIdAndUserId(ctx, entities.GetEventResultByEventIdAndUserIdParams{
		EventID: race.ID,
		UserID:  other.ID,
	}); err != nil {
		t.Fatalf("other result was removed: %v", err)
	}
	if _, err := events.GetEventByLocationAndTypeAndDate(ctx, entities.GetEventByLocationAndTypeAndDateParams{
		LocationID: track.ID,
		Type:       "Rental",
		Date:       "2024-06-01",
	}); err != nil {
		t.Fatalf("event was removed: %v", err)
	}
	remaining, err := friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   other.ID,
		Friendid: kept.ID,
	})
	if err != nil {
		t.Fatalf("kept friendship: %v", err)
	}
	if remaining.ID != keptFriend.ID {
		t.Fatalf("kept friendship id = %d, want %d", remaining.ID, keptFriend.ID)
	}
}

func TestGetMissingUser(t *testing.T) {
	repo := newUserRepo(t)
	ctx := context.Background()

	_, err := repo.GetUserById(ctx, 99)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get by id error = %v, want sql.ErrNoRows", err)
	}
	_, err = repo.GetUserByEmail(ctx, "missing@example.com")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get by email error = %v, want sql.ErrNoRows", err)
	}
	_, err = repo.GetUserByEmailForLogin(ctx, "missing@example.com")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get for login error = %v, want sql.ErrNoRows", err)
	}
	_, err = repo.GetUserIdByProfileId(ctx, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get by profile error = %v, want sql.ErrNoRows", err)
	}
}

func newUserRepo(t *testing.T) user.UserRepository {
	t.Helper()
	return user.NewUserRepository(entities.New(testdb.Open(t)))
}

func createUser(t *testing.T, repo user.UserRepository, firstName, lastName, email, profileID string) entities.GetUserByIdRow {
	t.Helper()
	ctx := context.Background()
	_, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Password:    "secret",
		ProfileID:   profileID,
		DisplayName: firstName,
	})
	if err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	byEmail, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user %s: %v", email, err)
	}
	byID, err := repo.GetUserById(ctx, byEmail.ID)
	if err != nil {
		t.Fatalf("get user id %d: %v", byEmail.ID, err)
	}
	return byID
}
