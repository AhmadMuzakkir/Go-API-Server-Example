package mock

import (
	"context"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
)

var _ store.UserStore = (*UserStore)(nil)

type UserStore struct {
	OnCreate func(ctx context.Context, username, passwordHash string) error
	OnGetByUsername func(ctx context.Context, username string) (*store.User, error)
	OnGetByID func(ctx context.Context, id int64) (*store.User, error)
}

func (u *UserStore) Create(ctx context.Context, username, passwordHash string) error {
	return u.OnCreate(ctx, username, passwordHash)
}

func (u *UserStore) GetByUsername(ctx context.Context, username string) (*store.User, error) {
	return u.OnGetByUsername(ctx, username)
}

func (u *UserStore) GetByID(ctx context.Context, id int64) (*store.User, error) {
	return u.OnGetByID(ctx, id)
}

