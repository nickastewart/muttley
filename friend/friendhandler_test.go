package friend_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"muttley/friend"
	"muttley/sqlite/entities"
	"muttley/sqlite/testdb"
	"muttley/user"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func TestAcceptFriendRequest(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")

	added := friendRequest(t, db, ada, http.MethodPost, "/add-friend?profileId=grace")
	if added.Code != http.StatusOK || !strings.Contains(added.Body.String(), "Pending Request") {
		t.Fatalf("add friend status = %d body = %s", added.Code, added.Body.String())
	}
	if status := friendStatus(t, db, ada.ID, grace.ID); status != "REQUESTED" {
		t.Fatalf("status after add = %q, want REQUESTED", status)
	}

	adaPage := friendRequest(t, db, ada, http.MethodGet, "/friends")
	if adaPage.Code != http.StatusOK {
		t.Fatalf("ada friends status = %d", adaPage.Code)
	}
	if strings.Contains(adaPage.Body.String(), "/accept-friend") || !strings.Contains(adaPage.Body.String(), `/remove-friend?profileId=grace`) {
		t.Fatalf("requester page = %s", adaPage.Body.String())
	}

	gracePage := friendRequest(t, db, grace, http.MethodGet, "/friends")
	if !strings.Contains(gracePage.Body.String(), `/accept-friend?profileId=ada`) || !strings.Contains(gracePage.Body.String(), `/remove-friend?profileId=ada`) {
		t.Fatalf("recipient page = %s", gracePage.Body.String())
	}

	selfAccept := friendRequest(t, db, ada, http.MethodPost, "/accept-friend?profileId=grace")
	if selfAccept.Code != http.StatusBadRequest {
		t.Fatalf("requester accept status = %d body = %s", selfAccept.Code, selfAccept.Body.String())
	}
	if status := friendStatus(t, db, ada.ID, grace.ID); status != "REQUESTED" {
		t.Fatalf("status after requester accept = %q, want REQUESTED", status)
	}

	accepted := friendRequest(t, db, grace, http.MethodPost, "/accept-friend?profileId=ada")
	if accepted.Code != http.StatusOK || !strings.Contains(accepted.Body.String(), `/remove-friend?profileId=ada`) || strings.Contains(accepted.Body.String(), "Accept") {
		t.Fatalf("accept status = %d body = %s", accepted.Code, accepted.Body.String())
	}
	if status := friendStatus(t, db, ada.ID, grace.ID); status != "ACCEPTED" {
		t.Fatalf("status after accept = %q, want ACCEPTED", status)
	}

	graceAfter := friendRequest(t, db, grace, http.MethodGet, "/friends")
	if strings.Contains(graceAfter.Body.String(), "/accept-friend") || !strings.Contains(graceAfter.Body.String(), `/remove-friend?profileId=ada`) {
		t.Fatalf("recipient page after accept = %s", graceAfter.Body.String())
	}

	again := friendRequest(t, db, grace, http.MethodPost, "/accept-friend?profileId=ada")
	if again.Code != http.StatusBadRequest {
		t.Fatalf("second accept status = %d body = %s", again.Code, again.Body.String())
	}
	if status := friendStatus(t, db, ada.ID, grace.ID); status != "ACCEPTED" {
		t.Fatalf("status after second accept = %q, want ACCEPTED", status)
	}
}

func TestSearchFriendsPendingRequest(t *testing.T) {
	db := testdb.Open(t)
	searcher := createUser(t, db, "Ada", "ada@example.com", "ada")
	match := createUser(t, db, "Grace", "grace@example.com", "grace")
	other := createUser(t, db, "Alan", "alan@example.com", "alan")

	friends := friend.NewFriendRepository(entities.New(db))
	if _, err := friends.AddFriend(context.Background(), entities.AddFriendParams{
		UserID:       searcher.ID,
		FriendID:     match.ID,
		FriendStatus: "REQUESTED",
	}); err != nil {
		t.Fatalf("add requested: %v", err)
	}
	if _, err := friends.AddFriend(context.Background(), entities.AddFriendParams{
		UserID:       other.ID,
		FriendID:     match.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add unrelated: %v", err)
	}

	res := friendRequest(t, db, searcher, http.MethodGet, "/search/friends?searchTerm=Grace")
	body := res.Body.String()
	if res.Code != http.StatusOK {
		t.Fatalf("search status = %d body = %s", res.Code, body)
	}
	if strings.Count(body, `class="friend-card"`) != 1 {
		t.Fatalf("result cards = %d body = %s", strings.Count(body, `class="friend-card"`), body)
	}
	if !strings.Contains(body, "Pending Request") || strings.Contains(body, `hx-post="/add-friend`) {
		t.Fatalf("search body = %s", body)
	}
}

func TestSearchFriendsBlankTermRestoresFriendsList(t *testing.T) {
	db := testdb.Open(t)
	searcher := createUser(t, db, "Ada", "ada@example.com", "ada")
	accepted := createUser(t, db, "Grace", "grace@example.com", "grace")
	if _, err := user.NewUserRepository(db).CreateUser(context.Background(), entities.CreateUserParams{
		FirstName:   "",
		LastName:    "",
		Email:       "blank@example.com",
		ProfileID:   "blank",
		DisplayName: "",
	}); err != nil {
		t.Fatalf("create blank user: %v", err)
	}
	if _, err := friend.NewFriendRepository(entities.New(db)).AddFriend(context.Background(), entities.AddFriendParams{
		UserID:       searcher.ID,
		FriendID:     accepted.ID,
		FriendStatus: "ACCEPTED",
	}); err != nil {
		t.Fatalf("add friend: %v", err)
	}

	paths := []string{
		"/search/friends?searchTerm=",
		"/search/friends",
		"/search/friends?searchTerm=%20%09%20",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			res := friendRequest(t, db, searcher, http.MethodGet, path)
			body := res.Body.String()
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d body = %s", res.Code, body)
			}
			if !strings.Contains(body, accepted.DisplayName) || !strings.Contains(body, ">Remove<") {
				t.Fatalf("friends list body = %s", body)
			}
			if strings.Contains(body, ">Add<") || strings.Contains(body, "blank") || strings.Contains(body, `class="friend-card-title"></div>`) || strings.Contains(body, `class="friend-card-title"> </div>`) {
				t.Fatalf("blank search rendered an empty add card: %s", body)
			}
			if strings.Count(body, `class="friend-card"`) != 1 {
				t.Fatalf("friend cards = %d body = %s", strings.Count(body, `class="friend-card"`), body)
			}
			if strings.Contains(body, `id="friend-list"`) || strings.Contains(body, "<html") {
				t.Fatalf("blank search rendered a full page: %s", body)
			}
		})
	}
}

func TestAcceptMissingFriendRequest(t *testing.T) {
	db := testdb.Open(t)
	ada := createUser(t, db, "Ada", "ada@example.com", "ada")
	grace := createUser(t, db, "Grace", "grace@example.com", "grace")

	missing := friendRequest(t, db, grace, http.MethodPost, "/accept-friend?profileId=ada")
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing request status = %d body = %s", missing.Code, missing.Body.String())
	}

	unknown := friendRequest(t, db, ada, http.MethodPost, "/accept-friend?profileId=nobody")
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("unknown user status = %d body = %s", unknown.Code, unknown.Body.String())
	}
}

func friendStatus(t *testing.T, db *sql.DB, userID, friendID int64) string {
	t.Helper()
	row, err := friend.NewFriendRepository(entities.New(db)).GetFriendByUserIdAndFriendId(context.Background(), entities.GetFriendByUserIdAndFriendIdParams{
		Userid:   userID,
		Friendid: friendID,
	})
	if err != nil {
		t.Fatalf("get friend: %v", err)
	}
	return row.FriendStatus
}

func friendRequest(t *testing.T, db *sql.DB, current entities.User, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := friend.NewFriendHandler(friend.NewFriendRepository(entities.New(db)), user.NewUserRepository(db))
	router := gin.New()
	router.HTMLRender = &testTemplRender{}
	asUser := func(c *gin.Context) {
		c.Set("currentUser", current)
		c.Next()
	}
	router.GET("/friends", asUser, handler.Friends)
	router.GET("/search/friends", asUser, handler.SearchFriends)
	router.POST("/add-friend", asUser, handler.AddFriend)
	router.POST("/accept-friend", asUser, handler.AcceptFriend)

	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

type testTemplRender struct {
	Code int
	Data templ.Component
}

func (t testTemplRender) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	w.WriteHeader(t.Code)
	if t.Data != nil {
		return t.Data.Render(context.Background(), w)
	}
	return nil
}

func (t testTemplRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (t *testTemplRender) Instance(name string, data interface{}) render.Render {
	if component, ok := data.(templ.Component); ok {
		return &testTemplRender{Code: http.StatusOK, Data: component}
	}
	return nil
}
