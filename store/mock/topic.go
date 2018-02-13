package mock

import (
	"github.com/disintegration/bebop/store"
)

// TopicStore is a mock implementation of store.TopicStore.
type TopicStore struct {
	OnNew           func(category, authorID int64, title string) (int64, error)
	OnGet           func(id int64) (*store.Topic, error)
	OnGetByCategory func(id int64, offset, limit int) ([]*store.Topic, int, error)
	OnSetTitle      func(id int64, title string) error
	OnDelete        func(id int64) error
}

func (s *TopicStore) New(category, authorID int64, title string) (int64, error) {
	return s.OnNew(category, authorID, title)
}
func (s *TopicStore) Get(id int64) (*store.Topic, error) {
	return s.OnGet(id)
}
func (s *TopicStore) GetByCategory(id int64, offset, limit int) ([]*store.Topic, int, error) {
	return s.OnGetByCategory(id, offset, limit)
}
func (s *TopicStore) SetTitle(id int64, title string) error {
	return s.OnSetTitle(id, title)
}
func (s *TopicStore) Delete(id int64) error {
	return s.OnDelete(id)
}
