package postgresql

import (
	"database/sql"
	"time"

	"github.com/disintegration/bebop/store"
)

type userStore struct {
	db *sql.DB
}

// New creates a new user.
func (s *userStore) New(authService string, authID string) (int64, error) {
	var id int64

	err := s.db.QueryRow(
		`insert into users(created_at, auth_service, auth_id) values($1, $2, $3) returning id`,
		time.Now(), authService, authID,
	).Scan(&id)

	return id, err
}

const selectFromUsers = `
	select 
		id, 
		coalesce(name, '') as name, 
		created_at,
		auth_service,
		auth_id,
		blocked,
		admin, 
		avatar 
	from users
`

func (s *userStore) scanUser(scanner scanner) (*store.User, error) {
	u := new(store.User)
	err := scanner.Scan(
		&u.ID,
		&u.Name,
		&u.CreatedAt,
		&u.AuthService,
		&u.AuthID,
		&u.Blocked,
		&u.Admin,
		&u.Avatar,
	)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Get finds a user by ID.
func (s *userStore) Get(id int64) (*store.User, error) {
	row := s.db.QueryRow(selectFromUsers+` where id=$1`, id)
	return s.scanUser(row)
}

// GetMany finds users by IDs.
func (s *userStore) GetMany(ids []int64) (map[int64]*store.User, error) {
	if len(ids) == 0 {
		return make(map[int64]*store.User), nil
	}

	users := make(map[int64]*store.User)
	for _, id := range ids {
		users[id] = nil
	}

	var params []interface{}
	for id := range users {
		params = append(params, id)
	}

	rows, err := s.db.Query(
		selectFromUsers+` where id in (`+placeholders(1, len(params))+`)`,
		params...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users[user.ID] = user
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, user := range users {
		if user == nil {
			return nil, store.ErrNotFound
		}
	}
	return users, nil
}

// GetAdmins finds all the admin users.
func (s *userStore) GetAdmins() ([]*store.User, error) {
	var users []*store.User

	rows, err := s.db.Query(selectFromUsers + ` where admin=true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetByName finds a user by name.
func (s *userStore) GetByName(name string) (*store.User, error) {
	row := s.db.QueryRow(selectFromUsers+` where name=$1`, name)
	return s.scanUser(row)
}

// GetByAuth finds a user by authService and authID.
func (s *userStore) GetByAuth(authService string, authID string) (*store.User, error) {
	row := s.db.QueryRow(selectFromUsers+` where auth_service=$1 and auth_id=$2`, authService, authID)
	return s.scanUser(row)
}

// SetName updates user.Name value. It returns ErrConflict if the given name is already taken.
func (s *userStore) SetName(id int64, name string) error {
	_, err := s.db.Exec(`update users set name=$1 where id=$2`, name, id)
	if isUniqueConstraintError(err) {
		return store.ErrConflict
	}
	return err
}

// SetBlocked updates user.Blocked value.
func (s *userStore) SetBlocked(id int64, blocked bool) error {
	_, err := s.db.Exec(`update users set blocked=$1 where id=$2`, blocked, id)
	return err
}

// SetAdmin updates user.Admin value.
func (s *userStore) SetAdmin(id int64, admin bool) error {
	_, err := s.db.Exec(`update users set admin=$1 where id=$2`, admin, id)
	return err
}

// SetAvatar updates user.Avatar value.
func (s *userStore) SetAvatar(id int64, avatar string) error {
	_, err := s.db.Exec(`update users set avatar=$1 where id=$2`, avatar, id)
	return err
}
