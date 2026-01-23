-- name: CreateUser :one
INSERT INTO users (name)
VALUES ($1)
RETURNING *;

-- name: GetUser :one
SELECT * 
FROM users
WHERE name = $1
LIMIT 1;

-- name: GetUserById :one
SELECT * 
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUsers :many
SELECT * 
FROM users;

-- name: Reset :exec
DELETE FROM users;
