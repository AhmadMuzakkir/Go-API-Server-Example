package mock

import (
	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
)

var _ store.Store = (*Store)(nil)

type Store struct {
	UserStore    store.UserStore
	MessageStore store.MessageStore
	TokenStore   store.TokenStore
}

func (s *Store) Message() store.MessageStore {
	return s.MessageStore
}

func (s *Store) Token() store.TokenStore {
	return s.TokenStore
}

func (s *Store) User() store.UserStore {
	return s.UserStore
}

