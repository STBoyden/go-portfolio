-- name: CreatePost :one
INSERT INTO "posts" (slug, content)
VALUES (@slug::text, @content::json)
RETURNING *;

-- name: GetPublishedPosts :many
SELECT * FROM "posts" WHERE published = TRUE;

-- name: GetPostByID :one
SELECT * FROM "posts" WHERE id = @id::uuid LIMIT 1;

-- name: GetAllPosts :many
SELECT * FROM "posts";
