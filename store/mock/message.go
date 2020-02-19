package mock

import (
	"context"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
)

var _ store.MessageStore = (*MessageStore)(nil)

type MessageStore struct {
	OnCreate  func(ctx context.Context, msg store.Message, recipientUserIDs []int64) error
	OnGet     func(ctx context.Context, userID int64) ([]*store.Message, error)
	OnGetByID func(ctx context.Context, msgID int64) (*store.Message, error)
	OnDelete  func(ctx context.Context, userID int64) error
	OnUpdate  func(ctx context.Context, msg store.Message, recipientUserIDs []int64) error
}

func (m *MessageStore) Create(ctx context.Context, msg store.Message, recipientUserIDs []int64) error {
	return m.OnCreate(ctx, msg, recipientUserIDs)
}

func (m *MessageStore) Get(ctx context.Context, userID int64) ([]*store.Message, error) {
	return m.OnGet(ctx, userID)
}

func (m *MessageStore) GetByID(ctx context.Context, msgID int64) (*store.Message, error) {
	return m.GetByID(ctx, msgID)
}

func (m *MessageStore) Delete(ctx context.Context, userID int64) error {
	return m.OnDelete(ctx, userID)
}

func (m *MessageStore) Update(ctx context.Context, msg store.Message, recipientUserIDs []int64) error {
	return m.OnUpdate(ctx, msg, recipientUserIDs)
}
