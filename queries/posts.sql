-- name: CreatePost :exec
INSERT INTO "posts" (slug, content) 
VALUES (@slug::text, @content::json);

-- name: GetPosts :many
SELECT * FROM "posts";