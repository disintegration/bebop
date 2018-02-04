package postgresql

import (
	"database/sql"
	"time"

	"github.com/disintegration/bebop/store"
)

type topicStore struct {
	db *sql.DB
}

// New creates a new topic.
func (s *topicStore) New(category, authorID int64, title string) (int64, error) {
	var id int64
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	err = s.db.QueryRow(
		`
			insert into topics(category_id, author_id, title, created_at, last_comment_at)
			values($1, $2, $3, $4)
			returning id
		`,
		category, authorID, title, now, now,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(`update categories set last_topic_at=$1, topic_count=topic_count+1 where id=$2`, now, category)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, nil
}

const selectFromTopics = `select id, category_id, author_id, title, created_at, last_comment_at, comment_count from topics`

func (s *topicStore) scanTopic(scanner scanner) (*store.Topic, error) {
	t := new(store.Topic)
	err := scanner.Scan(&t.ID, &t.CategoryID, &t.AuthorID, &t.Title, &t.CreatedAt, &t.LastCommentAt, &t.CommentCount)
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
	row := s.db.QueryRow(selectFromTopics+` where deleted=false and id=$1`, id)
	return s.scanTopic(row)
}

// GetByCategory returns a limited number of latest topics and a total topic count.
func (s *topicStore) GetByCategory(id int64, offset, limit int) ([]*store.Topic, int, error) {
	var count int
	err := s.db.QueryRow(`select count(*) from topics where (deleted=false and category_id=$1)`, id).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || offset > count {
		return []*store.Topic{}, count, nil
	}

	rows, err := s.db.Query(
		selectFromTopics+` where (deleted=false and category_id=$1) order by last_comment_at desc, id desc limit $2 offset $3`,
		id,
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
	_, err := s.db.Exec(`update topics set title=$1 where id=$2`, title, id)
	return err
}

// Delete soft-deletes a topic.
func (s *topicStore) Delete(id int64) error {
	_, err := s.db.Exec(`update topics set deleted=true where id=$1`, id)
	return err
}
