package mysql

import (
	"database/sql"
	"time"

	"github.com/disintegration/bebop/store"
)

type categoryStore struct {
	db *sql.DB
}

func (s *categoryStore) New(authorID int64, title string) (int64, error) {
	now := time.Now()

	res, err := s.db.Exec(
		`
			insert into categories(author_id, title, created_at, last_topic_at)
			values(?, ?, ?, ?)
		`,
		authorID, title, now, now,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

const selectFromCategories = `select id, author_id, title, created_at, last_topic_at, topic_count from categories`

func (s *categoryStore) scanCategory(scanner scanner) (*store.Category, error) {
	c := new(store.Category)
	err := scanner.Scan(&c.ID, &c.AuthorID, &c.Title, &c.CreatedAt, &c.LastTopicAt, &c.TopicCount)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Get finds a category by ID.
func (s *categoryStore) Get(id int64) (*store.Category, error) {
	row := s.db.QueryRow(selectFromCategories+` where deleted=false and id=?`, id)
	return s.scanCategory(row)
}

// GetLatest returns a limited number of latest categories and a total category count.
func (s *categoryStore) GetLatest(offset, limit int) ([]*store.Category, int, error) {
	var count int
	err := s.db.QueryRow(`select count(*) from categories where deleted=false`).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || offset > count {
		return []*store.Category{}, count, nil
	}

	rows, err := s.db.Query(
		selectFromCategories+` where deleted=false order by last_topic_at desc, id desc limit ? offset ?`,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	cats := []*store.Category{}
	for rows.Next() {
		cat, err := s.scanCategory(rows)
		if err != nil {
			return nil, 0, err
		}
		cats = append(cats, cat)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return cats, count, nil
}

// SetTitle updates category.Title value.
func (s *categoryStore) SetTitle(id int64, title string) error {
	_, err := s.db.Exec(`update categories set title=? where id=?`, title, id)
	return err
}

// Delete soft-deletes a category.
func (s *categoryStore) Delete(id int64) error {
	_, err := s.db.Exec(`update categories set deleted=true where id=?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`update topics set deleted=true where category_id=?`, id)
	return err
}
