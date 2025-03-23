// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: posts.sql

package persistence

import (
	"context"
)

const createPost = `-- name: CreatePost :exec
INSERT INTO "posts" (slug, content) 
VALUES ($1::text, $2::json)
`

type CreatePostParams struct {
	Slug    string `json:"slug"`
	Content []byte `json:"content"`
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) error {
	_, err := q.db.Exec(ctx, createPost, arg.Slug, arg.Content)
	return err
}

const getPosts = `-- name: GetPosts :many
SELECT id, created_at, slug, content FROM "posts"
`

func (q *Queries) GetPosts(ctx context.Context) ([]Post, error) {
	rows, err := q.db.Query(ctx, getPosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var i Post
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.Slug,
			&i.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
