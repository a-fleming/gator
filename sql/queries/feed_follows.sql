-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT
    i.id,
    i.created_at,
    i.updated_at,
    i.user_id,
    i.feed_id,
    u.name as user_name,
    f.name as feed_name
FROM
inserted i
JOIN users u ON u.id = i.user_id
JOIN feeds f ON f.id = i.feed_id;
