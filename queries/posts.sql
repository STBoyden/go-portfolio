-- name: CreatePost :one
INSERT INTO "posts" (id, slug, content)
VALUES (gen_random_uuid(), @slug::text, @content::json)
RETURNING *;

-- name: GetPublishedPosts :many
SELECT * FROM "posts" WHERE published = TRUE;

-- name: GetPostByID :one
SELECT * FROM "posts" WHERE id = @id::uuid LIMIT 1;

-- name: GetPostBySlug :one
SELECT * FROM "posts" WHERE slug = @slug::text LIMIT 1;

-- name: GetAllPosts :many
SELECT * FROM "posts";
