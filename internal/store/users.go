package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type UserStore struct {
	db *sql.DB
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

func (s *UserStore) Create(ctx context.Context, u *User) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := s.db.QueryRowContext(
		ctx,
		query,
		u.Username,
		u.Password,
		u.Email).
		Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) Get(ctx context.Context, id int64) (*User, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		SELECT id, username, email, password, created_at 
		FROM users 
		WHERE id = $1
	`
	var user User

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) Follow(ctx context.Context, followerID int64, userID int64) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)
	`
	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}
	return err
}

func (s *UserStore) UnFollow(ctx context.Context, followerID int64, userID int64) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		DELETE FROM followers WHERE follower_id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}
