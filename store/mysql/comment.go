package mysql

import (
	"database/sql"
	"time"

	"github.com/disintegration/bebop/store"
)

type commentStore struct {
	db *sql.DB
}

// New creates a new comment.
func (s *commentStore) New(topicID int64, authorID int64, content string) (int64, error) {
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	res, err := s.db.Exec(
		`insert into comments(topic_id, author_id, content, created_at) values (?, ?, ?, ?)`,
		topicID, authorID, content, now,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(`update topics set last_comment_at=?, comment_count=comment_count+1 where id=?`, now, topicID)
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

const selectFromComments = `select id, topic_id, author_id, content, created_at from comments`

func (s *commentStore) scanComment(scanner scanner) (*store.Comment, error) {
	c := new(store.Comment)
	err := scanner.Scan(&c.ID, &c.TopicID, &c.AuthorID, &c.Content, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Get finds a comment by ID.
func (s *commentStore) Get(id int64) (*store.Comment, error) {
	row := s.db.QueryRow(selectFromComments+` where deleted=false and id=?`, id)
	return s.scanComment(row)
}

// GetByTopic finds comments by topic.
func (s *commentStore) GetByTopic(topicID int64, offset, limit int) ([]*store.Comment, int, error) {
	var count int
	err := s.db.QueryRow(`select count(*) from comments where deleted=false and topic_id=?`, topicID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || offset > count {
		return []*store.Comment{}, count, nil
	}

	rows, err := s.db.Query(
		selectFromComments+` where deleted=false and topic_id=? order by created_at, id limit ? offset ?`,
		topicID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments := []*store.Comment{}
	for rows.Next() {
		comment, err := s.scanComment(rows)
		if err != nil {
			return nil, 0, err
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return comments, count, nil
}

// SetContent updates comment.Content value.
func (s *commentStore) SetContent(id int64, content string) error {
	_, err := s.db.Exec(
		`update comments set content=? where id=?`,
		content, id,
	)
	return err
}

// Delete soft-deletes a comment.
func (s *commentStore) Delete(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`update comments set deleted=true where id=?`, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		`
		update topics t set 
			last_comment_at=(select max(created_at) from comments c where c.topic_id=t.id and c.deleted=false),
			comment_count=t.comment_count-1
		where t.id=(select topic_id from comments where id = ?)
		`,
		id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
