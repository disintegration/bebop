package store

import (
	"time"
	"unicode/utf8"
)

// Topic is a discussion topic.
type Topic struct {
	ID            int64     `json:"id"`
	AuthorID      int64     `json:"authorId"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"createdAt"`
	LastCommentAt time.Time `json:"lastCommentAt"`
	CommentCount  int       `json:"commentCount"`
}

const (
	topicTitleMinLen = 1
	topicTitleMaxLen = 100
)

// ValidTopicTitle checks if topic title is valid.
func ValidTopicTitle(topicTitle string) bool {
	if !utf8.ValidString(topicTitle) {
		return false
	}

	length := utf8.RuneCountInString(topicTitle)
	if !(topicTitleMinLen <= length && length <= topicTitleMaxLen) {
		return false
	}

	return true
}
