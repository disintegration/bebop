package filestorage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local is a local file storage.
type Local struct {
	// dir is the base directory to store files.
	dir string

	// url is the base URL of the stored files.
	url string
}

// NewLocal returns a new local file storage.
func NewLocal(dir, url string) (*Local, error) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, fmt.Errorf("failed to create a directory (%q): %v", dir, err)
	}

	s := &Local{
		dir: dir,
		url: strings.TrimSuffix(url, "/"),
	}
	return s, nil
}

// Save saves data from r to file with the given path.
func (s *Local) Save(path string, r io.Reader) error {
	fullpath := filepath.Join(s.dir, path)

	if err := os.MkdirAll(filepath.Dir(fullpath), 0777); err != nil {
		return fmt.Errorf("failed to create a directory for file (%q): %v", fullpath, err)
	}

	w, err := os.Create(fullpath)
	if err != nil {
		return fmt.Errorf("failed to create a file (%q): %v", fullpath, err)
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to copy data to file (%q): %v", fullpath, err)
	}

	return nil
}

// Remove removes the file with the given path.
func (s *Local) Remove(path string) error {
	fullpath := filepath.Join(s.dir, path)

	if err := os.Remove(fullpath); err != nil {
		return fmt.Errorf("failed to remove a file (%q): %v", fullpath, err)
	}

	return nil
}

// URL returns an URL of the file with the given path.
func (s *Local) URL(path string) string {
	return s.url + "/" + path
}
