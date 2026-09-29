-- name: CreateEventResult :one
INSERT INTO event_result (event_id, user_id, best_lap_time, average_lap_time, position, number_of_laps) VALUES (?, ?, ?, ?, ?, ?)
    RETURNING *;

-- name: GetEventResultByEventIdAndUserId :one
SELECT * FROM event_result WHERE event_id = ? and user_id = ?;

-- name: GetUserFriendsResults :many
SELECT sqlc.embed(event), sqlc.embed(location), sqlc.embed(event_result), sqlc.embed(user) FROM event
    LEFT JOIN location ON event.location_id = location.id
    LEFT JOIN event_result ON event.id = event_result.event_id
    LEFT JOIN user ON user.id = event_result.user_id
    WHERE event_result.user_id in (SELECT friend_id FROM friend WHERE friend.user_id = ?);
