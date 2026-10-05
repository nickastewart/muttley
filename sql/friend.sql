-- name: AddFriend :one 
INSERT INTO friend (user_id, friend_id, friend_status) VALUES (?, ?, ?) RETURNING *;

-- name: GetFriendsByUser :many
SELECT user.id, user.first_name, user.last_name, user.profile_id, user.display_name,
       friend.friend_status AS friend_status,
       CASE
           WHEN friend.friend_id = sqlc.arg(userId)
                AND friend.friend_status = 'REQUESTED' THEN 'true'
           ELSE 'false'
       END AS confirmation_required
FROM user
JOIN friend ON (
        (friend.user_id = sqlc.arg(userId) AND friend.friend_id = user.id)
        OR
        (friend.friend_id = sqlc.arg(userId) AND friend.user_id = user.id)
    )
WHERE friend.friend_status != 'CANCELLED'
  AND user.id != sqlc.arg(userId);

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
