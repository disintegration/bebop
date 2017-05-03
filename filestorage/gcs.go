package filestorage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GoogleCloudStorage is a GCS-based file storage.
type GoogleCloudStorage struct {
	bucket string
	client *storage.Client
}

// NewGoogleCloudStorage returns a new GoogleCloudStorage file storage.
func NewGoogleCloudStorage(serviceAccountFile, bucket string) (*GoogleCloudStorage, error) {
	var opts []option.ClientOption
	if serviceAccountFile != "" {
		opts = append(opts, option.WithServiceAccountFile(serviceAccountFile))
	}

	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new GCS client: %v", err)
	}

	s := &GoogleCloudStorage{
		bucket: bucket,
		client: client,
	}
	return s, nil
}

// Save saves data from r to file with the given path.
func (s *GoogleCloudStorage) Save(path string, r io.Reader) error {
	w := s.client.Bucket(s.bucket).Object(path).NewWriter(context.Background())
	w.ACL = []storage.ACLRule{{
		Entity: storage.AllUsers,
		Role:   storage.RoleReader,
	}}
	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("failed to copy file to GCS bucket: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %v", err)
	}

	return nil
}

// Remove removes the file with the given path.
func (s *GoogleCloudStorage) Remove(path string) error {
	err := s.client.Bucket(s.bucket).Object(path).Delete(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete GCS object: %v", err)
	}
	return nil
}

// URL returns an URL of the file with the given path.
func (s *GoogleCloudStorage) URL(path string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, path)
}
