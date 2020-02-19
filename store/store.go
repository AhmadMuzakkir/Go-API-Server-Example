package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicate = errors.New("store: duplicate entry")
	ErrNotFound  = errors.New("store: item not found")
)

type Message struct {
	ID              int64     `json:"id"`
	Content         string    `json:"content"`
	SenderID        int64     `json:"-"`
	Sender          string    `json:"sender"`
	SentDateTime    time.Time `json:"sent_at"`
	UpdatedDateTime time.Time `json:"updated_at"`
}

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Token struct {
	UserID    int64
	UpdatedAt time.Time
}

type Store interface {
	Message() MessageStore
	User() UserStore
	Token() TokenStore
}

type MessageStore interface {
	Create(ctx context.Context, msg Message, recipientUserIDs []int64) error
	Get(ctx context.Context, userID int64) ([]*Message, error)
	GetByID(ctx context.Context, msgID int64) (*Message, error)
	Delete(ctx context.Context, userID int64) error
	Update(ctx context.Context, msg Message, recipientUserIDs []int64) error
}

type UserStore interface {
	Create(ctx context.Context, username, passwordHash string) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
}

type TokenStore interface {
	Create(ctx context.Context, userID int64, token string, updatedAt time.Time) error
	GetUserID(ctx context.Context, token string) (*Token, error)
}
