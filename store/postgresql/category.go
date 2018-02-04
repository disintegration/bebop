package postgresql

import (
	"database/sql"

	"github.com/disintegration/bebop/store"
)

type categoryStore struct {
	db *sql.DB
}

func (store *categoryStore) New(authorID int64, title string) (int64, error) {
	panic("not implemented")
}

func (store *categoryStore) Get(id int64) (*store.Category, error) {
	panic("not implemented")
}

func (store *categoryStore) GetLatest(offset, limit int) ([]*store.Category, int, error) {
	panic("not implemented")
}

func (store *categoryStore) SetTitle(id int64, title string) error {
	panic("not implemented")
}

func (store *categoryStore) Delete(id int64) error {
	panic("not implemented")
}
