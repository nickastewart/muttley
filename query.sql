-- name: GetUserById :one
SELECT id, first_name, last_name, email, created_at FROM user WHERE id = ?;

-- name: CreateUser :one
INSERT INTO user (first_name, last_name, email, password) VALUES (?, ?, ?, ?)
    RETURNING first_name, last_name, email, created_at;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email FROM user WHERE email = ?;

-- name: GetUserByEmailForLogin :one
SELECT id, email, password FROM user WHERE email = ?;

-- name: GetLocationByName :one 
SELECT id, name FROM location WHERE name = ?;

-- name: CreateLocation :one
INSERT INTO location (name) VALUES (?)
    RETURNING id, name;

-- name: GetEventByLocationAndTypeAndDate :one
SELECT * FROM event WHERE location_id = ? AND type = ? AND date = ?;

-- name: CreateEvent :one
INSERT INTO event (location_id, type, date, total_drivers) VALUES (?, ?, ?, ?)
    RETURNING *;

-- name: CreateEventResult :one
INSERT INTO event_result (event_id, user_id, best_lap_time, average_lap_time, position, number_of_laps) VALUES (?, ?, ?, ?, ?, ?)
    RETURNING *;

-- name: GetEventResultByEventIdAndUserId :one
SELECT * FROM event_result WHERE event_id = ? and user_id = ?; 

-- name: GetEventsByUser :many
SELECT sqlc.embed(event), sqlc.embed(location), sqlc.embed(event_result), sqlc.embed(user) FROM event
    LEFT JOIN location on event.location_id = location.id
    LEFT JOIN event_result on event.id = event_result.event_id
    LEFT JOIN user on user.id = event_result.user_id
    WHERE event_result.user_id = ?;

-- name: AddFriend :one 
INSERT INTO friend (user_id, friend_id, friend_status) VALUES (?, ?, ?) RETURNING *; 

-- name: GetFriendsByUser :many 
SELECT * FROM friend WHERE user_id = ? and friend_status = 'ACCEPTED'; 

-- name: GetUserFriendsResults :many
SELECT sqlc.embed(event), sqlc.embed(location), sqlc.embed(event_result), sqlc.embed(user) FROM event
    LEFT JOIN location on event.location_id = location.id
    LEFT JOIN event_result on event.id = event_result.event_id
    LEFT JOIN user on user.id = event_result.user_id
    WHERE event_result.user_id in (SELECT friend_id FROM friend WHERE friend.user_id = ?);
