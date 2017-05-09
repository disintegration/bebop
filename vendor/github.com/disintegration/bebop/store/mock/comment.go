package mock

import (
	"github.com/disintegration/bebop/store"
)

// CommentStore is a mock implementation of store.CommentStore.
type CommentStore struct {
	OnNew        func(topicID int64, authorID int64, content string) (int64, error)
	OnGet        func(id int64) (*store.Comment, error)
	OnGetByTopic func(topicID int64, offset, limit int) ([]*store.Comment, int, error)
	OnSetContent func(id int64, content string) error
	OnDelete     func(id int64) error
}

func (s *CommentStore) New(topicID int64, authorID int64, content string) (int64, error) {
	return s.OnNew(topicID, authorID, content)
}
func (s *CommentStore) Get(id int64) (*store.Comment, error) {
	return s.OnGet(id)
}
func (s *CommentStore) GetByTopic(topicID int64, offset, limit int) ([]*store.Comment, int, error) {
	return s.OnGetByTopic(topicID, offset, limit)
}
func (s *CommentStore) SetContent(id int64, content string) error {
	return s.OnSetContent(id, content)
}
func (s *CommentStore) Delete(id int64) error {
	return s.OnDelete(id)
}
