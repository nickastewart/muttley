-- name: GetDashboard :one
SELECT 
    COUNT(*) as totalRaces,
    CAST(COALESCE(CEIL(SUM(CASE WHEN position = 1 THEN 1 ELSE 0 END)), 0) AS INTEGER) as totalWins,
    CAST(COALESCE(ROUND(100.0 * SUM(CASE WHEN position = 1 THEN 1 ELSE 0 END) / COUNT(*), 1), 0) AS REAL) as winRate,
    CAST(COALESCE(CEIL(SUM(CASE WHEN position <= 3 THEN 1 ELSE 0 END)), 0) AS INTEGER) as totalPodiums,
    CAST(COALESCE(ROUND(100.0 * SUM(CASE WHEN position <= 3 THEN 1 ELSE 0 END) / COUNT(*), 1), 0) AS REAL) as podiumRate,
    CAST(COALESCE(CEIL(MIN(position)), 0) AS INTEGER) as bestPosition,
    CAST(COALESCE(CEIL(AVG(position)), 0) AS INTEGER) as avgPosition
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
