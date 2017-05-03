package mock

import (
	"github.com/disintegration/bebop/store"
)

// UserStore is a mock implementation of store.UserStore.
type UserStore struct {
	OnNew        func(authService string, authID string) (int64, error)
	OnGet        func(id int64) (*store.User, error)
	OnGetMany    func(ids []int64) (map[int64]*store.User, error)
	OnGetAdmins  func() ([]*store.User, error)
	OnGetByName  func(name string) (*store.User, error)
	OnGetByAuth  func(authService string, authID string) (*store.User, error)
	OnSetName    func(id int64, name string) error
	OnSetBlocked func(id int64, blocked bool) error
	OnSetAdmin   func(id int64, admin bool) error
	OnSetAvatar  func(id int64, avatar string) error
}

func (s *UserStore) New(authService string, authID string) (int64, error) {
	return s.OnNew(authService, authID)
}
func (s *UserStore) Get(id int64) (*store.User, error) {
	return s.OnGet(id)
}
func (s *UserStore) GetMany(ids []int64) (map[int64]*store.User, error) {
	return s.OnGetMany(ids)
}
func (s *UserStore) GetAdmins() ([]*store.User, error) {
	return s.OnGetAdmins()
}
func (s *UserStore) GetByName(name string) (*store.User, error) {
	return s.OnGetByName(name)
}
func (s *UserStore) GetByAuth(authService string, authID string) (*store.User, error) {
	return s.OnGetByAuth(authService, authID)
}
func (s *UserStore) SetName(id int64, name string) error {
	return s.OnSetName(id, name)
}
func (s *UserStore) SetBlocked(id int64, blocked bool) error {
	return s.OnSetBlocked(id, blocked)
}
func (s *UserStore) SetAdmin(id int64, admin bool) error {
	return s.OnSetAdmin(id, admin)
}
func (s *UserStore) SetAvatar(id int64, avatar string) error {
	return s.OnSetAvatar(id, avatar)
}
