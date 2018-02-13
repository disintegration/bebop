package postgresql

import (
	"reflect"
	"testing"

	"github.com/disintegration/bebop/store"
)

func TestCategory(t *testing.T) {
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

	cat2, err := s.Categories().New(u2, "cat-2")
	if err != nil {
		t.Fatalf("failed to create a category: %s", err)
	}

	cat3, err := s.Categories().New(u1, "cat-3")
	if err != nil {
		t.Fatalf("failed to create a category: %s", err)
	}

	cat4, err := s.Categories().New(u2, "cat-4 日本 Доброе утро")
	if err != nil {
		t.Fatalf("failed to create a category: %s", err)
	}

	cs, c, err := s.Categories().GetLatest(0, 10)
	if err != nil {
		t.Fatalf("failed to get latest categories: %s", err)
	}

	if len(cs) != 4 {
		t.Fatalf("bad categories len: %d", len(cs))
	}

	if c != 4 {
		t.Fatalf("bad categories count: %d", c)
	}

	for _, cat := range cs {
		if cat.ID != cat1 && cat.ID != cat2 && cat.ID != cat3 && cat.ID != cat4 {
			t.Fatalf("bad category id: got %d want one of (%d, %d, %d, %d)", cat.ID, cat1, cat2, cat3, cat4)
		}
	}

	cs, c, err = s.Categories().GetLatest(0, 2)
	if err != nil {
		t.Fatalf("failed to get all categories: %s", err)
	}

	if len(cs) != 2 {
		t.Fatalf("bad categories len: %d", len(cs))
	}

	if c != 4 {
		t.Fatalf("bad category count: %d", c)
	}

	got, err := s.Categories().Get(cat3)
	if err != nil {
		t.Fatalf("failed to get a category: %s", err)
	}

	want := &store.Category{
		ID:        cat3,
		AuthorID:  u1,
		Title:     "cat-3",
		CreatedAt: got.CreatedAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got topic %v, want %v", got, want)
	}

	err = s.Categories().SetTitle(cat3, "new title")
	if err != nil {
		t.Fatalf("failed to SetTitle: %s", err)
	}

	got, err = s.Categories().Get(cat3)
	if err != nil {
		t.Fatalf("failed to get a category: %s", err)
	}

	want = &store.Category{
		ID:        cat3,
		AuthorID:  u1,
		Title:     "new title",
		CreatedAt: got.CreatedAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got topic %v, want %v", got, want)
	}

	err = s.Categories().Delete(cat3)
	if err != nil {
		t.Fatalf("failed to delete category: %s", err)
	}

	_, err = s.Categories().Get(cat3)
	if err == nil {
		t.Fatal("expected error getting deleted category")
	}
}
