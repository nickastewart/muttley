-- name: GetEventByLocationAndTypeAndDate :one
SELECT * FROM event WHERE location_id = ? AND type = ? AND date = ?;

-- name: CreateEvent :one
INSERT INTO event (location_id, type, date, total_drivers) VALUES (?, ?, ?, ?)
    RETURNING *;

-- name: GetEventsByUser :many
SELECT sqlc.embed(event), sqlc.embed(location), sqlc.embed(event_result), sqlc.embed(user) FROM event
    LEFT JOIN location on event.location_id = location.id
    LEFT JOIN event_result on event.id = event_result.event_id
    LEFT JOIN user on user.id = event_result.user_id
    WHERE event_result.user_id in (sqlc.slice('ids'));

-- name: GetRecentEvents :many
SELECT sqlc.embed(event), sqlc.embed(event_result), sqlc.embed(location)
FROM event 
JOIN event_result on event.id = event_result.event_id
JOIN location on location.id = event.location_id
WHERE event_result.user_id = sqlc.arg(userId)
ORDER BY event.date DESC
LIMIT 10;
