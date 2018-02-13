package postgresql

import (
	"reflect"
	"testing"

	"github.com/disintegration/bebop/store"
)

func TestTopic(t *testing.T) {
	s, teardown := getTestStore(t)
	defer teardown()

	u1, err := s.Users().New("service1", "uid1")
	if err != nil {
		t.Fatalf("failed to create a user: %s", err)
	}
	u2, err := s.Users().New("service2", "uid2")
	if err != nil {
		t.Fatalf("failed to create a user: %s", err)
	}

	cat1, err := s.Categories().New(u1, "cat-1")
	if err != nil {
		t.Fatalf("failed to create a category: %s", err)
	}

	id1, err := s.Topics().New(cat1, u1, "topic1")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}
	id2, err := s.Topics().New(cat1, u2, "topic2")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}
	id3, err := s.Topics().New(cat1, u1, "topic3")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}
	id4, err := s.Topics().New(cat1, u2, "topic4 日本 Доброе утро")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}

	topics, c, err := s.Topics().GetByCategory(cat1, 0, 10)
	if err != nil {
		t.Fatalf("failed to get latest topics: %s", err)
	}

	if len(topics) != 4 {
		t.Fatalf("bad topics len: %d", len(topics))
	}

	if c != 4 {
		t.Fatalf("bad topic count: %d", c)
	}

	for _, topic := range topics {
		if topic.ID != id1 && topic.ID != id2 && topic.ID != id3 && topic.ID != id4 {
			t.Fatalf("bad topic id: got %d want one of (%d, %d, %d, %d)", topic.ID, id1, id2, id3, id4)
		}
	}

	topics, c, err = s.Topics().GetByCategory(cat1, 0, 2)
	if err != nil {
		t.Fatalf("failed to get all topics: %s", err)
	}

	if len(topics) != 2 {
		t.Fatalf("bad topics len: %d", len(topics))
	}

	if c != 4 {
		t.Fatalf("bad topic count: %d", c)
	}

	got, err := s.Topics().Get(id3)
	if err != nil {
		t.Fatalf("failed to get a topic: %s", err)
	}

	want := &store.Topic{
		ID:            id3,
		CategoryID:    cat1,
		AuthorID:      u1,
		Title:         "topic3",
		CreatedAt:     got.CreatedAt,
		LastCommentAt: got.LastCommentAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got topic %v, want %v", got, want)
	}

	err = s.Topics().SetTitle(id3, "new title")
	if err != nil {
		t.Fatalf("failed to SetTitle: %s", err)
	}

	got, err = s.Topics().Get(id3)
	if err != nil {
		t.Fatalf("failed to get a topic: %s", err)
	}

	want = &store.Topic{
		ID:            id3,
		CategoryID:    cat1,
		AuthorID:      u1,
		Title:         "new title",
		CreatedAt:     got.CreatedAt,
		LastCommentAt: got.LastCommentAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got topic %v, want %v", got, want)
	}

	err = s.Topics().Delete(id3)
	if err != nil {
		t.Fatalf("failed to delete topic: %s", err)
	}

	_, err = s.Topics().Get(id3)
	if err == nil {
		t.Fatal("expected error getting deleted topic")
	}
}
