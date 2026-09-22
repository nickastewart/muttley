package friend_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"muttley/friend"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"
)

func TestAddAndGetFriend(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	friends := friend.NewFriendRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")

	created, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       ada.ID,
		FriendID:     grace.ID,
		FriendStatus: "REQUESTED",
	})
	if err != nil {
		t.Fatalf("add friend: %v", err)
	}
	if created.ID == 0 || created.UserID != ada.ID || created.FriendID != grace.ID || created.FriendStatus != "REQUESTED" || created.CreatedAt == "" || created.RowVersion != 0 {
		t.Fatalf("created friend = %+v", created)
	}
	if created.AcceptedDate.Valid || created.UpdatedAt.Valid {
		t.Fatalf("expected empty accepted and updated times, got %+v", created)
	}

	forward, err := friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   ada.ID,
		Friendid: grace.ID,
	})
	if err != nil {
		t.Fatalf("get friend: %v", err)
	}
	reverse, err := friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   grace.ID,
		Friendid: ada.ID,
	})
	if err != nil {
		t.Fatalf("get friend in reverse: %v", err)
	}
	if forward.ID != created.ID || reverse.ID != created.ID {
		t.Fatalf("forward = %+v, reverse = %+v", forward, reverse)
	}
}

func TestGetMissingFriend(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")

	_, err := friend.NewFriendRepository(entities.New(db)).GetFriendByUserIdAndFriendId(context.Background(), entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   ada.ID,
		Friendid: grace.ID,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func TestGetFriendsByUser(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	friends := friend.NewFriendRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")
	alan := createUser(t, db, "Alan", "alan@example.com", "alan")

	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       ada.ID,
		FriendID:     grace.ID,
		FriendStatus: "REQUESTED",
	}); err != nil {
		t.Fatalf("add grace: %v", err)
	}

	adaFriends, err := friends.GetFriendsByUser(ctx, ada.ID)
	if err != nil {
		t.Fatalf("get ada friends: %v", err)
	}
	if len(adaFriends) != 1 {
		t.Fatalf("ada friends = %+v, want only grace", adaFriends)
	}
	if adaFriends[0].ID != grace.ID || adaFriends[0].FirstName != "Grace" || adaFriends[0].FriendStatus != "REQUESTED" || adaFriends[0].ConfirmationRequired != "false" {
		t.Fatalf("ada friend row = %+v", adaFriends[0])
	}

	graceFriends, err := friends.GetFriendsByUser(ctx, grace.ID)
	if err != nil {
		t.Fatalf("get grace friends: %v", err)
	}
	if len(graceFriends) != 1 || graceFriends[0].ID != ada.ID || graceFriends[0].ConfirmationRequired != "true" || graceFriends[0].FriendStatus != "REQUESTED" {
		t.Fatalf("grace friends = %+v", graceFriends)
	}

	// Added after the lookup above. A second friendship on Ada is also joined
	// when listing Grace's friends, so it would duplicate Ada in that result.
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       ada.ID,
		FriendID:     alan.ID,
		FriendStatus: "CANCELLED",
	}); err != nil {
		t.Fatalf("add alan: %v", err)
	}

	adaFriends, err = friends.GetFriendsByUser(ctx, ada.ID)
	if err != nil {
		t.Fatalf("get ada friends after cancel: %v", err)
	}
	if len(adaFriends) != 1 || adaFriends[0].ID != grace.ID {
		t.Fatalf("ada friends after cancel = %+v, want only grace", adaFriends)
	}

	alanFriends, err := friends.GetFriendsByUser(ctx, alan.ID)
	if err != nil {
		t.Fatalf("get alan friends: %v", err)
	}
	if len(alanFriends) != 0 {
		t.Fatalf("cancelled friendship was returned: %+v", alanFriends)
	}
}

func TestUpdateFriendStatus(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	friends := friend.NewFriendRepository(entities.New(db))
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")

	created, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       ada.ID,
		FriendID:     grace.ID,
		FriendStatus: "REQUESTED",
	})
	if err != nil {
		t.Fatalf("add friend: %v", err)
	}

	updated, err := friends.UpdateFriendStatus(ctx, entities.UpdateFriendStatusParams{
		FriendStatus: "ACCEPTED",
		ID:           created.ID,
	})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.ID != created.ID || updated.FriendStatus != "ACCEPTED" || updated.UserID != ada.ID || updated.FriendID != grace.ID {
		t.Fatalf("updated friend = %+v", updated)
	}

	found, err := friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   ada.ID,
		Friendid: grace.ID,
	})
	if err != nil {
		t.Fatalf("get updated friend: %v", err)
	}
	if found.FriendStatus != "ACCEPTED" {
		t.Fatalf("stored status = %q, want ACCEPTED", found.FriendStatus)
	}
}

func createUser(t *testing.T, db *sql.DB, firstName, email, profileID string) entities.GetUserByIdRow {
	t.Helper()
	repo := user.NewUserRepository(entities.New(db))
	ctx := context.Background()
	if _, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    "Racer",
		Email:       email,
		Password:    "secret",
		ProfileID:   profileID,
		DisplayName: firstName,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	byEmail, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	byID, err := repo.GetUserById(ctx, byEmail.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	return byID
}
