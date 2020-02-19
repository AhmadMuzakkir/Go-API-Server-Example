package mysql

import (
	"database/sql"
	"fmt"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	_ "github.com/go-sql-driver/mysql"
)

var _ store.Store = (*Store)(nil)

type Store struct {
	db *sql.DB

	messageStore *messageStore
	userStore    *userStore
	tokenStore   *tokenStore
}

func Connect(host string, port int, username, password, database string) (*Store, error) {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&multiStatements=True",
		username, password, host, port, database,
	)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{
		db:           db,
		messageStore: &messageStore{db: db},
		userStore:    &userStore{db: db},
		tokenStore:   &tokenStore{db: db},
	}

	return s, nil
}

func (s *Store) Message() store.MessageStore {
	return s.messageStore
}

func (s *Store) User() store.UserStore {
	return s.userStore
}

func (s *Store) Token() store.TokenStore {
	return s.tokenStore
}
