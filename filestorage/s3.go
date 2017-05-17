package filestorage

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AmazonS3 is an Amazon S3 file storage.
type AmazonS3 struct {
	bucket string
	svc    *s3.S3
}

// NewAmazonS3 returns a new Amazon S3 file storage.
func NewAmazonS3(accessKey, secretKey, region, bucket string) (*AmazonS3, error) {
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(region)

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	s := &AmazonS3{
		bucket: bucket,
		svc:    s3.New(sess),
	}
	return s, nil
}

// Save saves data from r to file with the given path.
func (s *AmazonS3) Save(path string, r io.Reader) error {
	_, err := s3manager.NewUploaderWithClient(s.svc).Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
			ACL:    aws.String(s3.ObjectCannedACLPublicRead),
			Body:   r,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload object to S3: %s", err)
	}
	return nil
}

// Remove removes the file with the given path.
func (s *AmazonS3) Remove(path string) error {
	_, err := s.svc.DeleteObject(
		&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(path),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %s", err)
	}
	return nil
}

// URL returns an URL of the file with the given path.
func (s *AmazonS3) URL(path string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, path)
}
