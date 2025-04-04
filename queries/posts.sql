-- name: CreatePost :one
INSERT INTO "posts" (id, slug, content)
VALUES (gen_random_uuid(), @slug::text, @content::json)
RETURNING *;

-- name: GetPublishedPosts :many
SELECT * FROM "posts" WHERE published = TRUE ORDER BY created_at DESC;

-- name: GetPublishedPostBySlug :one
SELECT * FROM "posts" WHERE published = TRUE AND slug = @slug::text LIMIT 1;

-- name: GetPostByID :one
SELECT * FROM "posts" WHERE id = @id::uuid LIMIT 1;

-- name: GetPostBySlug :one
SELECT * FROM "posts" WHERE slug = @slug::text LIMIT 1;

-- name: GetAllPosts :many
SELECT * FROM "posts" ORDER BY created_at DESC;

-- name: PublishPost :execrows
UPDATE "posts" SET published = TRUE WHERE id = @id::uuid;

-- name: UnpublishPost :execrows
UPDATE "posts" SET published = FALSE WHERE id = @id::uuid;

-- name: EditPost :execrows
UPDATE "posts"
SET content = @content::json
WHERE id = @id::uuid;
