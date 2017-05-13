package postgresql

import (
	"reflect"
	"testing"
	"time"

	"github.com/disintegration/bebop/store"
)

func TestUser(t *testing.T) {
	s, teardown := getTestStore(t)
	defer teardown()

	id, err := s.Users().New("service1", "user1")
	if err != nil {
		t.Fatalf("failed to create user1: %s", err)
	}
	_, err = s.Users().New("service2", "user2")
	if err != nil {
		t.Fatalf("failed to create user2: %s", err)
	}
	_, err = s.Users().New("service1", "user2")
	if err != nil {
		t.Fatalf("failed to create user3: %s", err)
	}

	_, err = s.Users().New("service1", "user2")
	if err == nil {
		t.Fatalf("expected error on duplicate auth")
	}

	user, err := s.Users().Get(id)
	if err != nil {
		t.Fatalf("failed to get user by id: %s", err)
	}

	sinceCreated := time.Since(user.CreatedAt)
	if sinceCreated > 3*time.Second || sinceCreated < 0 {
		t.Fatalf("bad user.CreatedAt: %v", user.CreatedAt)
	}

	want := &store.User{
		ID:          id,
		AuthService: "service1",
		AuthID:      "user1",
		CreatedAt:   user.CreatedAt,
	}

	if !reflect.DeepEqual(user, want) {
		t.Fatalf("got user %v want %v", user, want)
	}

	err = s.Users().SetAdmin(id, true)
	if err != nil {
		t.Fatalf("failed to SetAdmin: %s", err)
	}

	err = s.Users().SetAvatar(id, "avatar1")
	if err != nil {
		t.Fatalf("failed to SetAvatar: %s", err)
	}

	err = s.Users().SetName(id, "user1")
	if err != nil {
		t.Fatalf("failed to SetName: %s", err)
	}

	err = s.Users().SetBlocked(id, true)
	if err != nil {
		t.Fatalf("failed to SetBlocked: %s", err)
	}

	got, err := s.Users().Get(id)
	if err != nil {
		t.Fatalf("failed to get user by id: %s", err)
	}

	want = &store.User{
		ID:          id,
		AuthService: "service1",
		AuthID:      "user1",
		CreatedAt:   user.CreatedAt,
		Name:        "user1",
		Blocked:     true,
		Avatar:      "avatar1",
		Admin:       true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got user %v want %v", got, want)
	}

	got, err = s.Users().GetByName("user1")
	if err != nil {
		t.Fatalf("failed to get user by name: %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got user %v want %v", got, want)
	}
	user1 := got

	user2, err := s.Users().GetByAuth("service2", "user2")
	if err != nil {
		t.Fatalf("failed to get user by auth: %s", err)
	}

	admins, err := s.Users().GetAdmins()
	if err != nil {
		t.Fatalf("failed to get admins: %s", err)
	}

	if len(admins) != 1 || admins[0].ID != id {
		t.Fatalf("bad admin list: %#v", admins)
	}

	err = s.Users().SetName(user2.ID, "USER1")
	if err != store.ErrConflict {
		t.Fatalf("expected error ErrConflict on duplicate user name, got: %v", err)
	}

	users, err := s.Users().GetMany([]int64{user.ID, user2.ID})
	if err != nil {
		t.Fatalf("failed to get many users by ids: %s", err)
	}

	if !reflect.DeepEqual(users[user1.ID], user1) {
		t.Fatalf("got user %v want %v", users[user1.ID], user1)
	}
	if !reflect.DeepEqual(users[user2.ID], user2) {
		t.Fatalf("got user %v want %v", users[user2.ID], user2)
	}
}
