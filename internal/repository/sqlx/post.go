package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"myfacebook/internal/db"
	"myfacebook/internal/repository"
)

type PostRepository struct {
	writeDB *db.DB
	readDB  *db.DB
}

func NewPostRepository(writeDB, readDB *db.DB) *PostRepository {
	return &PostRepository{
		writeDB: writeDB,
		readDB:  readDB,
	}
}

func (r *PostRepository) Add(ctx context.Context, post repository.Post) error {
	dbConn := r.writeDB.GetConnection()

	sqlQuery := `INSERT INTO posts (id, text, author_id) 
				VALUES (:id, :text, :author_id)`

	_, err := dbConn.NamedExecContext(ctx, sqlQuery, post)
	if err != nil {
		return fmt.Errorf("failed to add post to db: %w", err)
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, postID, authorID string) error {
	dbConn := r.writeDB.GetConnection()

	_, err := dbConn.ExecContext(ctx, `DELETE FROM posts WHERE id=$1 AND author_id=$2`, postID, authorID)
	if err != nil {
		return fmt.Errorf("failed to delete post from db: %w", err)
	}

	return nil
}

func (r *PostRepository) GetPostByID(ctx context.Context, postID string) (*repository.Post, error) {
	dbConn := r.readDB.GetConnection()

	var post repository.Post

	sqlQuery := `SELECT id, text, author_id FROM posts WHERE id = $1`

	err := dbConn.GetContext(ctx, &post, sqlQuery, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	return &post, nil
}

func (r *PostRepository) GetPostsByIDs(ctx context.Context, postIDs []string, offset, limit int) ([]repository.Post, error) {
	dbConn := r.readDB.GetConnection()

	var posts []repository.Post

	sqlQuery, args, err := sqlx.In(`SELECT id, text, author_id FROM posts WHERE id IN (?) ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		postIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare sql IN query: %w", err)
	}

	sqlQuery = dbConn.Rebind(sqlQuery)

	err = dbConn.SelectContext(ctx, &posts, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by ids: %w", err)
	}

	return posts, nil
}

func (r *PostRepository) GetLastPostsIDsByAuthorIDs(ctx context.Context, authorIDs []string, createdAfterTimestampMilli int64, limit int) ([]string, error) {
	dbConn := r.readDB.GetConnection()

	var postsIDs []string

	sqlQuery := `SELECT id FROM posts WHERE author_id IN (?)`

	var args []interface{}
	args = append(args, authorIDs)

	if createdAfterTimestampMilli > 0 {
		sqlQuery += ` AND EXTRACT(EPOCH FROM created_at AT TIME ZONE 'UTC') * 1000 >= ?` // convert to milli

		args = append(args, createdAfterTimestampMilli)
	}

	sqlQuery += ` ORDER BY created_at DESC LIMIT ?`

	args = append(args, limit)

	sqlQuery, args2, err := sqlx.In(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare sql IN query: %w", err)
	}

	sqlQuery = dbConn.Rebind(sqlQuery)

	err = dbConn.SelectContext(ctx, &postsIDs, sqlQuery, args2...)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts ids by author ids: %w", err)
	}

	return postsIDs, nil
}

func (r *PostRepository) Update(ctx context.Context, post repository.Post) error {
	dbConn := r.writeDB.GetConnection()

	sqlQuery := `UPDATE posts SET text=$2 WHERE id=$1`

	res, err := dbConn.ExecContext(ctx, sqlQuery, post.ID, post.Text)
	if err != nil {
		return fmt.Errorf("failed to update post in db: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
