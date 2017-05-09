// Package static provides http handlers for serving static files
// from the embedded filesystem or from a local directory.
package static

//go:generate go run gen.go

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"time"
)

// Dir returns a handler that serves static files from the given local directory.
// Unlike http.FileServer(http.Dir(path)) it will not serve directory listings.
func Dir(stripPrefix string, path string) http.Handler {
	handler := http.FileServer(&dir{Dir: http.Dir(path)})
	if stripPrefix != "" {
		handler = http.StripPrefix(stripPrefix, handler)
	}
	return handler
}

type dir struct {
	http.Dir
}

func (d *dir) Open(name string) (http.File, error) {
	f, err := d.Dir.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, os.ErrNotExist
	}
	return f, nil
}

// Embedded is an http handler that serves static files embedded into source of this package.
func Embedded(stripPrefix string) http.Handler {
	handler := http.FileServer(fs)
	if stripPrefix != "" {
		handler = http.StripPrefix(stripPrefix, handler)
	}
	return handler
}

// EmbeddedFile is an http handler that serves a static file embedded into source of this package.
func EmbeddedFile(file string) http.Handler {
	handler := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = file
		handler.ServeHTTP(w, r)
	})
}

type embeddedFilesystem map[string]*fileData

func (efs embeddedFilesystem) Open(name string) (http.File, error) {
	if fd, ok := efs[name]; ok {
		f := &embeddedFile{
			Reader:   bytes.NewReader(fd.body),
			fileData: fd,
		}
		return f, nil
	}
	return nil, os.ErrNotExist
}

type fileData struct {
	name  string
	size  int64
	mtime int64
	body  []byte
}

type embeddedFile struct {
	*bytes.Reader
	*fileData
}

func (f *embeddedFile) Close() error {
	return nil
}
func (f *embeddedFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *embeddedFile) Stat() (os.FileInfo, error) {
	return f, nil
}
func (f *embeddedFile) Name() string {
	return f.name
}

func (f *embeddedFile) Size() int64 {
	return f.size
}

func (f *embeddedFile) Mode() os.FileMode {
	return 0
}

func (f *embeddedFile) ModTime() time.Time {
	return time.Unix(f.mtime, 0)
}

func (f *embeddedFile) IsDir() bool {
	return false
}

func (f *embeddedFile) Sys() interface{} {
	return nil
}
