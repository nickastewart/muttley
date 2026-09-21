-- name: GetUserById :one
SELECT id, first_name, last_name, email, profile_id, display_name, created_at FROM user WHERE id = ?;

-- name: CreateUser :one
INSERT INTO user (first_name, last_name, email, password, profile_id, display_name) VALUES (?, ?, ?, ?, ?, ?)
    RETURNING first_name, last_name, email, profile_id, display_name, created_at;

-- name: ResetPassword :exec 
UPDATE user SET password = ? WHERE email = ?;

-- name: UpdateUser :exec
UPDATE user SET first_name = ?, last_name = ?, email = ?, display_name = ? WHERE id = ?;

-- name: DeleteEventResultsByUserId :exec
DELETE FROM event_result WHERE user_id = ?;

-- name: DeleteFriendsByUserId :exec
DELETE FROM friend WHERE user_id = sqlc.arg(userId) OR friend_id = sqlc.arg(userId);

-- name: DeleteUser :exec
DELETE FROM user WHERE id = ?; 

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

-- name: GetDashboard :one
SELECT 
    COUNT(*) as totalRaces,
    CEIL(SUM(CASE WHEN position = 1 THEN 1 ELSE 0 END)) as totalWins,
    ROUND(100.0 * SUM(CASE WHEN position = 1 THEN 1 ELSE 0 END) / COUNT(*), 1) as winRate,
    CEIL(SUM(CASE WHEN position <= 3 THEN 1 ELSE 0 END)) as totalPodiums,
    ROUND(100.0 * SUM(CASE WHEN position <= 3 THEN 1 ELSE 0 END) / COUNT(*), 1) as podiumRate,
    CEIL(MIN(position)) as bestPosition,
    CEIL(AVG(position)) as avgPosition
FROM event_result
WHERE user_id = sqlc.arg(userId);

-- name: GetBestTrack :one 
WITH location_stats AS (
    SELECT 
        location.name,
        AVG(event_result.position) AS avgPosition,
        COUNT(*) as totalRaces
    FROM event_result
    JOIN event ON event.id = event_result.event_id
    JOIN location ON location.id = event.location_id
    WHERE event_result.user_id = sqlc.arg(userId)
    GROUP BY location.name
    HAVING COUNT(*) >= 3) 
SELECT 
    name, 
    CEIL(avgPosition) AS avgPosition 
FROM location_stats ORDER BY avgPosition ASC LIMIT 1;

-- name: GetLocationStats :many 
SELECT location.name, COUNT(*) FROM event 
JOIN location ON event.location_id = location.id
JOIN event_result ON event.id = event_result.event_id
WHERE event_result.user_id = sqlc.arg(userId)
GROUP BY location.name
ORDER BY COUNT(*) DESC 
LIMIT 3;

-- name: GetRecentPositions :many 
SELECT position FROM event_result
JOIN event ON event_result.event_id = event.id
WHERE user_id = sqlc.arg(userId) 
ORDER BY event.date DESC
LIMIT 20;

-- name: GetRecentEvents :many
SELECT sqlc.embed(event), sqlc.embed(event_result), sqlc.embed(location)
FROM event 
JOIN event_result on event.id = event_result.event_id
JOIN location on location.id = event.location_id
WHERE event_result.user_id = sqlc.arg(userId)
ORDER BY event.date DESC
LIMIT 10;
