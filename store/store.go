// Package store provides a bebop data store interface.
package store

import (
	"errors"
)

var (
	// ErrNotFound means the requested item is not found.
	ErrNotFound = errors.New("store: item not found")
	// ErrConflict means the operation failed because of a conflict between items.
	ErrConflict = errors.New("store: item conflict")
)

// Store is a bebop data store interface.
type Store interface {
	Users() UserStore
	Topics() TopicStore
	Comments() CommentStore
}

// UserStore is a bebop user data store interface.
type UserStore interface {
	New(authService string, authID string) (int64, error)
	Get(id int64) (*User, error)
	GetMany(ids []int64) (map[int64]*User, error)
	GetAdmins() ([]*User, error)
	GetByName(name string) (*User, error)
	GetByAuth(authService string, authID string) (*User, error)
	SetName(id int64, name string) error
	SetBlocked(id int64, blocked bool) error
	SetAdmin(id int64, admin bool) error
	SetAvatar(id int64, avatar string) error
}

// TopicStore is a bebop topic data store interface.
type TopicStore interface {
	New(authorID int64, title string) (int64, error)
	Get(id int64) (*Topic, error)
	GetLatest(offset, limit int) ([]*Topic, int, error)
	SetTitle(id int64, title string) error
	Delete(id int64) error
}

// CommentStore is a bebop comment data store interface.
type CommentStore interface {
	New(topicID int64, authorID int64, content string) (int64, error)
	Get(id int64) (*Comment, error)
	GetByTopic(topicID int64, offset, limit int) ([]*Comment, int, error)
	SetContent(id int64, content string) error
	Delete(id int64) error
}
