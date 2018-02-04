package mock

import "github.com/disintegration/bebop/store"

// CategoryStore is a mock implementation of store.CategoryStore.
type CategoryStore struct {
	OnNew       func(authorID int64, title string) (int64, error)
	OnGet       func(id int64) (*store.Category, error)
	OnGetLatest func(offset, limit int) ([]*store.Category, int, error)
	OnSetTitle  func(id int64, title string) error
	OnDelete    func(id int64) error
}

func (s *CategoryStore) New(authorID int64, title string) (int64, error) {
	return s.OnNew(authorID, title)
}

func (s *CategoryStore) Get(id int64) (*store.Category, error) {
	return s.OnGet(id)
}

func (s *CategoryStore) GetLatest(offset, limit int) ([]*store.Category, int, error) {
	return s.OnGetLatest(offset, limit)
}

func (s *CategoryStore) SetTitle(id int64, title string) error {
	return s.OnSetTitle(id, title)
}

func (s *CategoryStore) Delete(id int64) error {
	return s.OnDelete(id)
}
