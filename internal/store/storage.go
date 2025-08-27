package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrConflict         = errors.New("resource alreay exists")
	QueryTimeoutDuraton = 5 * time.Second
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Update(context.Context, *Post) error
		Delete(context.Context, int64) error
	}
	Users interface {
		Create(context.Context, *User) error
		Get(context.Context, int64) (*User, error)
		Follow(context.Context, int64, int64) error
		UnFollow(context.Context, int64, int64) error
	}
	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) (*Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
	}
}
