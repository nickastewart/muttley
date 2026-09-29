-- name: GetHeadToHead :many
SELECT
    friend_user.id,
    friend_user.first_name,
    friend_user.last_name,
    friend_user.display_name,
    CAST(COALESCE(SUM(CASE WHEN mine.position < theirs.position THEN 1 ELSE 0 END), 0) AS INTEGER) AS user_wins,
    CAST(COALESCE(SUM(CASE WHEN theirs.position < mine.position THEN 1 ELSE 0 END), 0) AS INTEGER) AS friend_wins,
    CAST(COUNT(*) AS INTEGER) AS shared_races
FROM event_result AS mine
JOIN event_result AS theirs
    ON theirs.event_id = mine.event_id
    AND theirs.user_id != mine.user_id
JOIN user AS friend_user
    ON friend_user.id = theirs.user_id
WHERE mine.user_id = sqlc.arg(userId)
    AND theirs.user_id IN (
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
GROUP BY friend_user.id, friend_user.first_name, friend_user.last_name, friend_user.display_name
ORDER BY shared_races DESC, friend_user.display_name COLLATE NOCASE ASC, friend_user.id ASC;
