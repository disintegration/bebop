package store

import (
	"time"
	"unicode/utf8"
)

// User represents an authenticated user.
// Only public fields are marshalled to JSON by default.
type User struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	AuthService string    `json:"-"`
	AuthID      string    `json:"-"`
	Blocked     bool      `json:"-"`
	Admin       bool      `json:"-"`
	Avatar      string    `json:"avatar"`
}

const (
	userNameMinLen = 3
	userNameMaxLen = 20
)

// validUserNameRune checks if given user name rune is valid.
func validUserNameRune(r rune) bool {
	if 'a' <= r && r <= 'z' {
		return true
	}
	if 'A' <= r && r <= 'Z' {
		return true
	}
	if '0' <= r && r <= '9' {
		return true
	}
	if r == '_' || r == '-' {
		return true
	}
	return false
}

// ValidUserName checks if given user name is valid.
func ValidUserName(userName string) bool {
	length := utf8.RuneCountInString(userName)
	if !(userNameMinLen <= length && length <= userNameMaxLen) {
		return false
	}

	for _, r := range userName {
		if !validUserNameRune(r) {
			return false
		}
	}

	return true
}
