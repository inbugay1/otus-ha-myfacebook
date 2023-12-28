package repository

import (
	"context"
)

type Post struct {
	ID       string `db:"id"`
	Text     string `db:"text"`
	AuthorID string `db:"author_id"`
}

type PostRepository interface {
	Add(ctx context.Context, post Post) error
	Delete(ctx context.Context, postID, authorID string) error
	GetPostByID(ctx context.Context, postID string) (*Post, error)
	GetPostsByIDs(ctx context.Context, postIDs []string, offset, limit int) ([]Post, error)
	GetLastPostsIDsByAuthorIDs(ctx context.Context, authorIDs []string, createdAfterTimestamp int64, limit int) ([]string, error)
	Update(ctx context.Context, post Post) error
}
