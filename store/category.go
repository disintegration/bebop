package store

import (
	"time"
	"unicode/utf8"
)

// Category is a general theme of topics.
type Category struct {
	ID          int64     `json:"id"`
	AuthorID    int64     `json:"authorId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"createdAt"`
	LastTopicAt time.Time `json:"lastTopicAt"`
	TopicCount  int       `json:"topicCount"`
}

const (
	categoryTitleMinLen = 1
	categoryTitleMaxLen = 100
)

// ValidCategoryTitle checks if a category title is valid.
func ValidCategoryTitle(title string) bool {
	if !utf8.ValidString(title) {
		return false
	}

	length := utf8.RuneCountInString(title)
	if !(categoryTitleMinLen <= length && length <= categoryTitleMaxLen) {
		return false
	}

	return true
}
