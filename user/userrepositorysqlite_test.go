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
}

func TestSearchUsersByName(t *testing.T) {
	db := testdb.Open(t)
	repo := user.NewUserRepository(db)
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

func TestSearchUsersByNameUsesSearcherFriendship(t *testing.T) {
	db := testdb.Open(t)
	repo := user.NewUserRepository(db)
	friends := friend.NewFriendRepository(entities.New(db))
	ctx := context.Background()

	searcher := createUser(t, repo, "Nick", "Stewart", "nick@example.com", "nick")
	pending := createUser(t, repo, "Ada", "Lovelace", "ada@example.com", "ada")
	other := createUser(t, repo, "Grace", "Hopper", "grace@example.com", "grace")
	cancelled := createUser(t, repo, "Alan", "Turing", "alan@example.com", "alan")

	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       searcher.ID,
		FriendID:     pending.ID,
		FriendStatus: "REQUESTED",
	}); err != nil {
		t.Fatalf("add requested: %v", err)
	}
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       other.ID,
		FriendID:     pending.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add unrelated: %v", err)
	}
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       searcher.ID,
		FriendID:     cancelled.ID,
		FriendStatus: "CANCELLED",
	}); err != nil {
		t.Fatalf("add cancelled: %v", err)
	}

	pendingMatches, err := repo.GetUsersBySearchTerm(ctx, entities.GetUsersBySearchTermParams{
		Name:   "%ada%",
		Userid: searcher.ID,
	})
	if err != nil {
		t.Fatalf("search ada: %v", err)
	}
	if len(pendingMatches) != 1 || pendingMatches[0].ID != pending.ID || pendingMatches[0].FriendStatus != "REQUESTED" {
		t.Fatalf("ada matches = %+v", pendingMatches)
	}

	cancelledMatches, err := repo.GetUsersBySearchTerm(ctx, entities.GetUsersBySearchTermParams{
		Name:   "%turing%",
		Userid: searcher.ID,
	})
	if err != nil {
		t.Fatalf("search turing: %v", err)
	}
	if len(cancelledMatches) != 1 || cancelledMatches[0].ID != cancelled.ID || cancelledMatches[0].FriendStatus != "NONE" {
		t.Fatalf("turing matches = %+v", cancelledMatches)
	}
}

func TestDeleteUserRemovesResultsAndFriendships(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	users := user.NewUserRepository(db)
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

func TestDeleteUserRemovesMagicLinks(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	users := user.NewUserRepository(db)

	owner := createUser(t, users, "Ada", "Lovelace", "ada@example.com", "ada")
	insertMagicLink(t, db, "ada@example.com", "unused-hash", sql.NullString{})
	insertMagicLink(t, db, "ada@example.com", "used-hash", sql.NullString{String: "2024-01-01T00:00:00Z", Valid: true})
	otherID := insertMagicLink(t, db, "grace@example.com", "other-hash", sql.NullString{})

	if err := users.DeleteUser(ctx, owner.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err := users.GetUserById(ctx, owner.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted user error = %v, want sql.ErrNoRows", err)
	}

	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM magic_link WHERE email = ?`, owner.Email).Scan(&remaining); err != nil {
		t.Fatalf("count deleted email links: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("magic links for deleted email = %d, want 0", remaining)
	}

	var keptEmail string
	if err := db.QueryRow(`SELECT email FROM magic_link WHERE id = ?`, otherID).Scan(&keptEmail); err != nil {
		t.Fatalf("kept magic link: %v", err)
	}
	if keptEmail != "grace@example.com" {
		t.Fatalf("kept magic link email = %s, want grace@example.com", keptEmail)
	}
}

func TestDeleteUserRollsBackWhenUserDeleteFails(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()
	users := user.NewUserRepository(db)
	locations := location.NewLocationRepository(entities.New(db))
	events := event.NewEventRepository(entities.New(db))
	results := eventresult.NewEventResultRepository(entities.New(db))
	friends := friend.NewFriendRepository(entities.New(db))

	owner := createUser(t, users, "Ada", "Lovelace", "ada@example.com", "ada")
	other := createUser(t, users, "Grace", "Hopper", "grace@example.com", "grace")

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
		t.Fatalf("create result: %v", err)
	}
	if _, err := friends.AddFriend(ctx, entities.AddFriendParams{
		UserID:       owner.ID,
		FriendID:     other.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add friend: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TRIGGER fail_user_delete BEFORE DELETE ON user
		BEGIN
			SELECT RAISE(ABORT, 'forced failure');
		END`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	if err := users.DeleteUser(ctx, owner.ID); err == nil {
		t.Fatal("expected delete to fail")
	}

	if _, err := users.GetUserById(ctx, owner.ID); err != nil {
		t.Fatalf("user was removed: %v", err)
	}
	if _, err := results.GetEventResultByEventIdAndUserId(ctx, entities.GetEventResultByEventIdAndUserIdParams{
		EventID: race.ID,
		UserID:  owner.ID,
	}); err != nil {
		t.Fatalf("result was removed: %v", err)
	}
	if _, err := friends.GetFriendByUserIdAndFriendId(ctx, entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   owner.ID,
		Friendid: other.ID,
	}); err != nil {
		t.Fatalf("friendship was removed: %v", err)
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
	_, err = repo.GetUserIdByProfileId(ctx, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get by profile error = %v, want sql.ErrNoRows", err)
	}
}

func insertMagicLink(t *testing.T, db *sql.DB, email, tokenHash string, usedAt sql.NullString) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO magic_link (email, token_hash, purpose, expires_at, used_at)
		VALUES (?, ?, ?, ?, ?)`,
		email, tokenHash, "login", "2099-01-01T00:00:00Z", usedAt,
	)
	if err != nil {
		t.Fatalf("insert magic link %s: %v", tokenHash, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("magic link id %s: %v", tokenHash, err)
	}
	return id
}

func newUserRepo(t *testing.T) user.UserRepository {
	t.Helper()
	return user.NewUserRepository(testdb.Open(t))
}

func createUser(t *testing.T, repo user.UserRepository, firstName, lastName, email, profileID string) entities.User {
	t.Helper()
	ctx := context.Background()
	_, err := repo.CreateUser(ctx, entities.CreateUserParams{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
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
