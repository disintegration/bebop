package filestorage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// AmazonS3 is an Amazon S3 file storage.
type AmazonS3 struct {
	region string
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
		region: region,
		bucket: bucket,
		svc:    s3.New(sess),
	}

	return s, nil
}

// Save saves data from r to file with the given path.
func (s *AmazonS3) Save(path string, r io.Reader) error {
	var body io.ReadSeeker
	if rs, ok := r.(io.ReadSeeker); ok {
		body = rs
	} else {
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read from r: %s", err)
		}
		body = bytes.NewReader(buf)
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   body,
	}

	if _, err := s.svc.PutObject(params); err != nil {
		return fmt.Errorf("failed to put object to s3: %s", err)
	}

	return nil
}

// Remove removes the file with the given path.
func (s *AmazonS3) Remove(path string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	if _, err := s.svc.DeleteObject(params); err != nil {
		return fmt.Errorf("failed to delete object from s3: %s", err)
	}

	return nil
}

// URL returns an URL of the file with the given path.
func (s *AmazonS3) URL(path string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, path)
}
