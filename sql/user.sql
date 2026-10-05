-- name: GetUserById :one
SELECT id, first_name, last_name, email, profile_id, display_name, created_at FROM user WHERE id = ?;

-- name: CreateUser :one
INSERT INTO user (first_name, last_name, email, profile_id, display_name) VALUES (?, ?, ?, ?, ?)
    RETURNING first_name, last_name, email, profile_id, display_name, created_at;

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

-- name: GetUsersBySearchTerm :many
SELECT user.id, user.first_name, user.last_name, user.profile_id,
       CAST(COALESCE(
           (SELECT friend.friend_status
            FROM friend
            WHERE friend.friend_status != 'CANCELLED'
              AND (
                    (friend.user_id = sqlc.arg(userId) AND friend.friend_id = user.id)
                    OR
                    (friend.friend_id = sqlc.arg(userId) AND friend.user_id = user.id)
                  )
            LIMIT 1),
           'NONE'
       ) AS TEXT) AS friend_status
FROM user
WHERE CONCAT(LOWER(user.first_name), ' ', LOWER(user.last_name)) LIKE sqlc.arg(name)
  AND user.id != sqlc.arg(userId);

-- name: GetUserIdByProfileId :one
SELECT user.id FROM user WHERE user.profile_id = ?;
