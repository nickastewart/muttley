package event_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"muttley/event"
	"muttley/sqlite/entities"

	"github.com/gin-gonic/gin"
)

func TestLeaderboardReportsAFriendQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/leaderboard", func(c *gin.Context) {
		c.Set("currentUser", entities.User{ID: 1})
		event.NewEventsHandler(nil, nil, nil, nil, failingFriends{}).Leaderboard(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	router.ServeHTTP(recorder, request)

	body := recorder.Body.String()
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(body, "Error getting friends") {
		t.Fatalf("status = %d body = %s", recorder.Code, body)
	}
	if strings.Contains(body, "database closed") || strings.Contains(body, "SELECT") {
		t.Fatalf("body leaked the query error: %s", body)
	}
	if recorder.Header().Get("HX-Redirect") != "" {
		t.Fatalf("redirect = %q", recorder.Header().Get("HX-Redirect"))
	}
}

type failingFriends struct{}

func (failingFriends) GetFriendsByUser(context.Context, int64) ([]entities.GetFriendsByUserRow, error) {
	return nil, errors.New("database closed: SELECT id FROM friend")
}

func (failingFriends) AddFriend(context.Context, entities.AddFriendParams) (entities.Friend, error) {
	return entities.Friend{}, errors.New("database closed: SELECT id FROM friend")
}

func (failingFriends) UpdateFriendStatus(context.Context, entities.UpdateFriendStatusParams) (entities.Friend, error) {
	return entities.Friend{}, errors.New("database closed: SELECT id FROM friend")
}

func (failingFriends) GetFriendByUserIdAndFriendId(context.Context, entities.GetFriendByUserIdAndFriendIdParams) (entities.Friend, error) {
	return entities.Friend{}, errors.New("database closed: SELECT id FROM friend")
}
