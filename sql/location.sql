-- name: GetLocationByName :one 
SELECT id, name FROM location WHERE name = ?;

-- name: CreateLocation :one
INSERT INTO location (name) VALUES (?)
    RETURNING id, name;
