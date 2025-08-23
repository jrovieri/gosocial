package store

import (
	"context"
	"database/sql"
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

func (s *UserStore) Create(ctx context.Context, u *User) error {
	query := `INSERTO INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at`
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
