// Package postgres provides a PostgreSQL implementation of the bebop data store interface.
package postgres

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/disintegration/bebop/store"
)

// Store is a postgres implementation of store.
type Store struct {
	db           *sql.DB
	userStore    *userStore
	topicStore   *topicStore
	commentStore *commentStore
}

// Users returns a user store.
func (s *Store) Users() store.UserStore {
	return s.userStore
}

// Topics returns a topic store.
func (s *Store) Topics() store.TopicStore {
	return s.topicStore
}

// Comments returns a comment store.
func (s *Store) Comments() store.CommentStore {
	return s.commentStore
}

var _ store.Store = (*Store)(nil)

// Connect connects to a store.
func Connect(address, username, password, database string) (*Store, error) {
	connstr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		username, password, address, database,
	)

	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{
		db:           db,
		userStore:    &userStore{db: db},
		topicStore:   &topicStore{db: db},
		commentStore: &commentStore{db: db},
	}

	err = s.Migrate()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Migrate migrates the store database.
func (s *Store) Migrate() error {
	for _, q := range migrate {
		_, err := s.db.Exec(q)
		if err != nil {
			return fmt.Errorf("sql exec error: %s; query: %q", err, q)
		}
	}
	return nil
}

// Drop drops the store database.
func (s *Store) Drop() error {
	for _, q := range drop {
		_, err := s.db.Exec(q)
		if err != nil {
			return fmt.Errorf("sql exec error: %s; query: %q", err, q)
		}
	}
	return nil
}

// Reset resets the store database.
func (s *Store) Reset() error {
	err := s.Drop()
	if err != nil {
		return err
	}
	return s.Migrate()
}

type scanner interface {
	Scan(v ...interface{}) error
}

func placeholders(start, count int) string {
	buf := new(bytes.Buffer)
	for i := start; i < start+count; i++ {
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(i))
		if i < start+count-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}
