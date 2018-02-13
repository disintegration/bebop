package mock

import (
	"github.com/disintegration/bebop/store"
)

// Store is a mock implementation of store.Store.
type Store struct {
	UserStore     *UserStore
	CategoryStore *CategoryStore
	TopicStore    *TopicStore
	CommentStore  *CommentStore
}

func (s *Store) Users() store.UserStore {
	return s.UserStore
}
func (s *Store) Categories() store.CategoryStore {
	return s.CategoryStore
}
func (s *Store) Topics() store.TopicStore {
	return s.TopicStore
}
func (s *Store) Comments() store.CommentStore {
	return s.CommentStore
}

var _ store.Store = (*Store)(nil)
