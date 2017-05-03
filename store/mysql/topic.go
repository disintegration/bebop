package mysql

import (
	"database/sql"
	"time"

	"github.com/disintegration/bebop/store"
)

type topicStore struct {
	db *sql.DB
}

// New creates a new topic.
func (s *topicStore) New(authorID int64, title string) (int64, error) {
	now := time.Now()

	res, err := s.db.Exec(
		`
			insert into topics(author_id, title, created_at, last_comment_at)
			values(?, ?, ?, ?)
		`,
		authorID, title, now, now,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

const selectFromTopics = `select id, author_id, title, created_at, last_comment_at, comment_count from topics`

func (s *topicStore) scanTopic(scanner scanner) (*store.Topic, error) {
	t := new(store.Topic)
	err := scanner.Scan(&t.ID, &t.AuthorID, &t.Title, &t.CreatedAt, &t.LastCommentAt, &t.CommentCount)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Get finds a topic by ID.
func (s *topicStore) Get(id int64) (*store.Topic, error) {
	row := s.db.QueryRow(selectFromTopics+` where deleted=false and id=?`, id)
	return s.scanTopic(row)
}

// GetLatest returns a limited number of latest topics and a total topic count.
func (s *topicStore) GetLatest(offset, limit int) ([]*store.Topic, int, error) {
	var count int
	err := s.db.QueryRow(`select count(*) from topics where deleted=false`).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || offset > count {
		return []*store.Topic{}, count, nil
	}

	rows, err := s.db.Query(
		selectFromTopics+` where deleted=false order by last_comment_at desc, id desc limit ? offset ?`,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	topics := []*store.Topic{}
	for rows.Next() {
		topic, err := s.scanTopic(rows)
		if err != nil {
			return nil, 0, err
		}
		topics = append(topics, topic)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return topics, count, nil
}

// SetTitle updates topic.Title value.
func (s *topicStore) SetTitle(id int64, title string) error {
	_, err := s.db.Exec(`update topics set title=? where id=?`, title, id)
	return err
}

// Delete soft-deletes a topic.
func (s *topicStore) Delete(id int64) error {
	_, err := s.db.Exec(`update topics set deleted=true where id=?`, id)
	return err
}
