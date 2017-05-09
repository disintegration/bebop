package mysql

import (
	"testing"
)

func TestComment(t *testing.T) {
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

	t1, err := s.Topics().New(u1, "topic1")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}
	t2, err := s.Topics().New(u2, "topic2")
	if err != nil {
		t.Fatalf("failed to create a topic: %s", err)
	}

	c1, err := s.Comments().New(t1, u1, "comment1")
	if err != nil {
		t.Fatalf("failed to create a comment: %s", err)
	}
	c2, err := s.Comments().New(t1, u2, "comment2")
	if err != nil {
		t.Fatalf("failed to create a comment: %s", err)
	}
	c3, err := s.Comments().New(t2, u1, "comment3")
	if err != nil {
		t.Fatalf("failed to create a comment: %s", err)
	}
	c4, err := s.Comments().New(t2, u2, "comment4 日本 Доброе утро")
	if err != nil {
		t.Fatalf("failed to create a comment: %s", err)
	}

	topic1, err := s.Topics().Get(t1)
	if err != nil {
		t.Fatalf("failed to get a topic: %s", err)
	}
	if topic1.CommentCount != 2 {
		t.Fatalf("bad topic1.CommentCount: %d", topic1.CommentCount)
	}

	err = s.Comments().Delete(c2)
	if err != nil {
		t.Fatalf("failed to delete a comment: %s", err)
	}

	_, err = s.Comments().Get(c2)
	if err == nil {
		t.Fatal("expected error getting deleted comment")
	}

	topic1, err = s.Topics().Get(t1)
	if err != nil {
		t.Fatalf("failed to get a topic: %s", err)
	}
	if topic1.CommentCount != 1 {
		t.Fatalf("bad topic1.CommentCount: %d", topic1.CommentCount)
	}

	comment1, err := s.Comments().Get(c1)
	if err != nil {
		t.Fatalf("failed to get a comment: %s", err)
	}
	if comment1.Content != "comment1" {
		t.Fatalf("bad comment content: %s", comment1.Content)
	}

	err = s.Comments().SetContent(c1, "new content")
	if err != nil {
		t.Fatalf("failed to SetContent: %s", err)
	}

	comment1, err = s.Comments().Get(c1)
	if err != nil {
		t.Fatalf("failed to get a comment: %s", err)
	}
	if comment1.Content != "new content" {
		t.Fatalf("bad comment content: %s", comment1.Content)
	}

	comments, count, err := s.Comments().GetByTopic(c2, 0, 10)
	if err != nil {
		t.Fatalf("failed to get comments by topic: %s", err)
	}

	if len(comments) != 2 {
		t.Fatalf("bad comments len: %d", len(comments))
	}
	if count != 2 {
		t.Fatalf("bad comment count: %d", count)
	}
	if comments[0].ID != c3 || comments[1].ID != c4 {
		t.Fatalf("bad comment ids: got (%d, %d) want (%d, %d)", comments[0].ID, comments[1].ID, c3, c4)
	}

	comments, count, err = s.Comments().GetByTopic(c2, 10, 1)
	if err != nil {
		t.Fatalf("failed to get comments by topic: %s", err)
	}

	if len(comments) != 0 {
		t.Fatalf("bad comments len: %d", len(comments))
	}
	if count != 2 {
		t.Fatalf("bad comment count: %d", count)
	}
}
