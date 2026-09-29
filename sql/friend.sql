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

-- name: UpdateFriendStatus :one 
UPDATE friend SET friend_status = ? WHERE id = ? RETURNING *;
