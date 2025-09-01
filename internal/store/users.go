package store

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type UserStore struct {
	db *sql.DB
}

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email,omitempty"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at,omitempty"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash
	return nil
}

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := tx.QueryRowContext(
		ctx,
		query,
		u.Username,
		u.Password.hash,
		u.Email).
		Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
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

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true

		if err := s.updateUserActivation(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string,
	invitationExp time.Duration, userID int64) error {

	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(invitationExp))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
			JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha512.Sum512([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) updateUserActivation(ctx context.Context, tx *sql.Tx, user *User) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `
		UPDATE users 
			SET username = $1, email = $2, is_active = $3 
		WHERE id = $4`

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuraton)
	defer cancel()

	query := `DELETE FROM user_invitations WHERE user_id = $1`

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
