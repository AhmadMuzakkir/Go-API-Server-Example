package mock

import (
	"context"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
)

var _ store.TokenStore = (*TokenStore)(nil)

type TokenStore struct {
	OnCreate    func(ctx context.Context, userID int64, token string, updatedAt time.Time) error
	OnGetUserID func(ctx context.Context, token string) (*store.Token, error)
}

func (t *TokenStore) Create(ctx context.Context, userID int64, token string, updatedAt time.Time) error {
	return t.OnCreate(ctx, userID, token, updatedAt)
}

func (t *TokenStore) GetUserID(ctx context.Context, token string) (*store.Token, error) {
	return t.OnGetUserID(ctx, token)
}
