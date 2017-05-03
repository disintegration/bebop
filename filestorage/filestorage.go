// Package filestorage provides a file storage service that manages
// public web app files e.g. user-uploaded avatars.
package filestorage

import (
	"io"
)

// FileStorage manages files uploaded by users.
type FileStorage interface {
	// Save saves data from r to file with the given path.
	Save(path string, r io.Reader) error

	// Remove removes the file with the given path.
	Remove(path string) error

	// URL returns an URL of the file with the given path.
	URL(path string) string
}
