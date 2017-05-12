package avatar

import (
	"github.com/disintegration/bebop/store"
)

// MockService is a mock implementation of avatar.Service
type MockService struct {
	OnSave     func(user *store.User, imageData []byte) error
	OnGenerate func(user *store.User) error
	OnURL      func(user *store.User) string
}

func (s *MockService) Save(user *store.User, imageData []byte) error {
	return s.OnSave(user, imageData)
}
func (s *MockService) Generate(user *store.User) error {
	return s.OnGenerate(user)
}
func (s *MockService) URL(user *store.User) string {
	return s.OnURL(user)
}
