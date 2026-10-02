-- name: ListLocationRecords :many
WITH user_locations AS (
    SELECT
        location.id AS location_id,
        location.name AS location_name,
        MAX(event.date) AS latest_result
    FROM event_result
    JOIN event ON event.id = event_result.event_id
    JOIN location ON location.id = event.location_id
    WHERE event_result.user_id = sqlc.arg(userId)
    GROUP BY location.id, location.name
),
personal_best AS (
    SELECT location_id, best_lap_time, average_lap_time, event_date
    FROM (
        SELECT
            event.location_id AS location_id,
            event_result.best_lap_time AS best_lap_time,
            event_result.average_lap_time AS average_lap_time,
            event.date AS event_date,
            ROW_NUMBER() OVER (
                PARTITION BY event.location_id
                ORDER BY event_result.best_lap_time ASC, event.date ASC, event.id ASC, event_result.id ASC
            ) AS rn
        FROM event_result
        JOIN event ON event.id = event_result.event_id
        WHERE event_result.user_id = sqlc.arg(userId)
            AND event_result.best_lap_time > 0
    ) AS ranked_personal_best
    WHERE rn = 1
),
best_average AS (
    SELECT location_id, average_lap_time, event_date
    FROM (
        SELECT
            event.location_id AS location_id,
            event_result.average_lap_time AS average_lap_time,
            event.date AS event_date,
            ROW_NUMBER() OVER (
                PARTITION BY event.location_id
                ORDER BY event_result.average_lap_time ASC, event.date ASC, event.id ASC, event_result.id ASC
            ) AS rn
        FROM event_result
        JOIN event ON event.id = event_result.event_id
        WHERE event_result.user_id = sqlc.arg(userId)
            AND event_result.average_lap_time > 0
    ) AS ranked_best_average
    WHERE rn = 1
),
track_record AS (
    SELECT
        location_id,
        best_lap_time,
        holder_id,
        display_name,
        first_name,
        last_name
    FROM (
        SELECT
            event.location_id AS location_id,
            event_result.best_lap_time AS best_lap_time,
            user.id AS holder_id,
            user.display_name AS display_name,
            user.first_name AS first_name,
            user.last_name AS last_name,
            ROW_NUMBER() OVER (
                PARTITION BY event.location_id
                ORDER BY event_result.best_lap_time ASC, event.date ASC, event.id ASC, event_result.id ASC
            ) AS rn
        FROM event_result
        JOIN event ON event.id = event_result.event_id
        JOIN user ON user.id = event_result.user_id
        WHERE event_result.best_lap_time > 0
            AND event_result.user_id IN (
                SELECT sqlc.arg(userId)
                UNION
                SELECT friend.friend_id
                FROM friend
                WHERE friend.user_id = sqlc.arg(userId)
                    AND friend.friend_status = 'ACCEPTED'
                UNION
                SELECT friend.user_id
                FROM friend
                WHERE friend.friend_id = sqlc.arg(userId)
                    AND friend.friend_status = 'ACCEPTED'
            )
    ) AS ranked_track_record
    WHERE rn = 1
)
SELECT
    user_locations.location_id,
    user_locations.location_name,
    CAST(COALESCE(personal_best.best_lap_time, 0) AS INTEGER) AS personal_best_lap,
    COALESCE(personal_best.event_date, '') AS personal_best_date,
    CAST(COALESCE(personal_best.average_lap_time, 0) AS INTEGER) AS personal_best_average_lap,
    CAST(COALESCE(best_average.average_lap_time, 0) AS INTEGER) AS best_average_lap,
    COALESCE(best_average.event_date, '') AS best_average_date,
    CAST(COALESCE(track_record.best_lap_time, 0) AS INTEGER) AS track_record_lap,
    COALESCE(track_record.display_name, '') AS track_record_display_name,
    COALESCE(track_record.first_name, '') AS track_record_first_name,
    COALESCE(track_record.last_name, '') AS track_record_last_name
FROM user_locations
LEFT JOIN personal_best ON personal_best.location_id = user_locations.location_id
LEFT JOIN best_average ON best_average.location_id = user_locations.location_id
LEFT JOIN track_record ON track_record.location_id = user_locations.location_id
ORDER BY user_locations.latest_result DESC, user_locations.location_name COLLATE NOCASE ASC;

-- name: GetLocationRecordSnapshot :one
SELECT
    CAST(COALESCE((
        SELECT MIN(event_result.best_lap_time)
        FROM event_result
        JOIN event ON event.id = event_result.event_id
        WHERE event_result.user_id = sqlc.arg(userId)
            AND event.location_id = sqlc.arg(locationId)
            AND event_result.best_lap_time > 0
    ), 0) AS INTEGER) AS personal_best_lap,
    CAST(COALESCE((
        SELECT MIN(event_result.best_lap_time)
        FROM event_result
        JOIN event ON event.id = event_result.event_id
        WHERE event.location_id = sqlc.arg(locationId)
            AND event_result.best_lap_time > 0
            AND event_result.user_id IN (
                SELECT sqlc.arg(userId)
                UNION
                SELECT friend.friend_id
                FROM friend
                WHERE friend.user_id = sqlc.arg(userId)
                    AND friend.friend_status = 'ACCEPTED'
                UNION
                SELECT friend.user_id
                FROM friend
                WHERE friend.friend_id = sqlc.arg(userId)
                    AND friend.friend_status = 'ACCEPTED'
            )
    ), 0) AS INTEGER) AS track_record_lap;
