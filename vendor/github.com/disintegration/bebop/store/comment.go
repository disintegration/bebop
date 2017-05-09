package store

import (
	"time"
	"unicode/utf8"
)

// Comment is a single comment on a topic.
type Comment struct {
	ID        int64     `json:"id"`
	TopicID   int64     `json:"topicId"`
	AuthorID  int64     `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	commentContentMinLen = 1
	commentContentMaxLen = 10000
)

// ValidCommentContent checks if comment content is valid.
func ValidCommentContent(commentContent string) bool {
	if !utf8.ValidString(commentContent) {
		return false
	}

	length := utf8.RuneCountInString(commentContent)
	if !(commentContentMinLen <= length && length <= commentContentMaxLen) {
		return false
	}

	return true
}
