-- name: GetUserById :one
SELECT id, first_name, last_name, email, profile_id, created_at FROM user WHERE id = ?;

-- name: CreateUser :one
INSERT INTO user (first_name, last_name, email, password, profile_id, display_name) VALUES (?, ?, ?, ?, ?, ?)
    RETURNING first_name, last_name, email, profile_id, display_name, created_at;

-- name: ResetPassword :exec 
UPDATE user SET password = ? WHERE email = ?; 

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email, profile_id FROM user WHERE email = ?;

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
    WHERE event_result.user_id in (sqlc.slice('ids'));

-- name: AddFriend :one 
INSERT INTO friend (user_id, friend_id, friend_status) VALUES (?, ?, ?) RETURNING *; 

-- name: GetFriendsByUser :many 
SELECT user.id, user.first_name, user.last_name, user.profile_id, user.display_name, COALESCE(friend.friend_status, 'NONE') AS friend_status,
        CASE 
            WHEN friend.friend_id = sqlc.arg(userId) AND COALESCE(friend.friend_status, 'NONE') = 'REQUESTED' THEN 'true'
            ELSE 'false'
        END AS confirmation_required
    FROM user
    LEFT JOIN friend ON (user.id = friend.user_id OR user.id = friend.friend_id)
    WHERE user.id in (
        SELECT friend.user_id FROM friend WHERE friend.friend_id = sqlc.arg(userId) AND friend_status != 'CANCELLED'
        UNION
        SELECT friend.friend_id FROM friend WHERE friend.user_id = sqlc.arg(userId) AND friend_status != 'CANCELLED'
    ) AND user.id != sqlc.arg(userId); 

-- name: GetFriendByUserIdAndFriendId :one 
SELECT *
FROM friend
WHERE 
    (friend_id = sqlc.arg(userId) AND user_id = sqlc.arg(friendId))
    OR
    (user_id = sqlc.arg(userId) AND friend_id = sqlc.arg(friendId))
LIMIT 1;
    
-- name: GetUserFriendsResults :many
SELECT sqlc.embed(event), sqlc.embed(location), sqlc.embed(event_result), sqlc.embed(user) FROM event
    LEFT JOIN location ON event.location_id = location.id
    LEFT JOIN event_result ON event.id = event_result.event_id
    LEFT JOIN user ON user.id = event_result.user_id
    WHERE event_result.user_id in (SELECT friend_id FROM friend WHERE friend.user_id = ?);

-- name: GetUsersBySearchTerm :many
SELECT user.id, user.first_name, user.last_name, user.profile_id, COALESCE(f1.friend_status, f2.friend_status, 'NONE') as friend_status FROM user 
    LEFT JOIN friend f1 ON f1.user_id = user.id
    LEFT JOIN friend f2 ON f2.friend_id = user.id
    WHERE CONCAT(LOWER(user.first_name), ' ', LOWER(user.last_name)) LIKE sqlc.arg(name) AND user.id != sqlc.arg(userId);

-- name: GetUserIdByProfileId :one
SELECT user.id FROM user WHERE user.profile_id = ?;

-- name: UpdateFriendStatus :one 
UPDATE friend SET friend_status = ? WHERE id = ? RETURNING *;
