-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeeds :many
SELECT *
FROM feeds;

-- name: GetFeedByUrl :one
SELECT * 
FROM feeds
WHERE url = $1
LIMIT 1;