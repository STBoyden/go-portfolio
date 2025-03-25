-- name: CreatePost :one
INSERT INTO "posts" (slug, content) 
VALUES (@slug::text, @content::json)
RETURNING *;

-- name: GetPosts :many
SELECT * FROM "posts";