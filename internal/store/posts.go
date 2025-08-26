package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, p *Post) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `INSERT INTO posts (content, title, user_id, tags) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	err := s.db.QueryRowContext(
		ctx,
		query,
		p.Content,
		p.Title,
		p.UserID,
		pq.Array(p.Tags)).
		Scan(
			&p.ID,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `SELECT id, user_id, title, content, tags, created_at, updated_at, version 
		FROM posts WHERE id = $1`

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		UPDATE posts SET title = $1, content = $2, version = version + 1 
			WHERE id = $3 AND version = $4 
			RETURNING version
	`
	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).
		Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrResourceNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) Delete(ctx context.Context, postID int64) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `DELETE FROM posts WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrResourceNotFound
	}
	return nil
}
