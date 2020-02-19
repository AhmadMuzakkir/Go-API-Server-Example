package mysql

import (
	"context"
	"database/sql"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/go-sql-driver/mysql"
)

var _ store.UserStore = (*userStore)(nil)

type userStore struct {
	db *sql.DB
}

func (s *userStore) Create(ctx context.Context, username, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users(username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok {
			if sqlErr.Number == 1062 {
				return store.ErrDuplicate
			}
		}
		return err
	}

	return nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*store.User, error) {
	row := s.db.QueryRow("SELECT id, username, password_hash FROM users WHERE username=?", username)

	var u store.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *userStore) GetByID(ctx context.Context, id int64) (*store.User, error) {
	row := s.db.QueryRow("SELECT id, username, password_hash FROM users WHERE id=?", id)

	var u store.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &u, nil
}
