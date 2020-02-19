package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
)

var _ store.TokenStore = (*tokenStore)(nil)

type tokenStore struct {
	db *sql.DB
}

func (t *tokenStore) Create(ctx context.Context, userID int64, token string, updatedAt time.Time) error {
	_, err := t.db.ExecContext(ctx, "INSERT INTO tokens(user_id, token, updated_at) VALUES(?,?,?) ON DUPLICATE KEY UPDATE token=?, updated_at=?", userID, token, updatedAt, token, updatedAt)
	return err
}

func (t *tokenStore) GetUserID(ctx context.Context, userToken string) (*store.Token, error) {
	row := t.db.QueryRowContext(ctx, "SELECT user_id, updated_at FROM tokens WHERE token=?", userToken)

	var token store.Token

	err := row.Scan(&token.UserID, &token.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &token, nil
}
