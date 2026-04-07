package media

import (
	"errors"
	"io"
)

// S3Storage is an optional backend placeholder for S3-compatible storage.
// A real implementation would use github.com/aws/aws-sdk-go-v2/service/s3
type S3Storage struct {
	BucketName string
	Region     string
	Endpoint   string
}

func NewS3Storage() *S3Storage {
	// Dummy initialization
	return &S3Storage{}
}

func (s *S3Storage) Save(name string, content io.Reader) (string, error) {
	return "", errors.New("S3Storage is not implemented. Please install the AWS SDK and configure.")
}

func (s *S3Storage) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("S3Storage is not implemented.")
}

func (s *S3Storage) Delete(name string) error {
	return errors.New("S3Storage is not implemented.")
}

func (s *S3Storage) Exists(name string) bool {
	return false
}

func (s *S3Storage) URL(name string) string {
	return ""
}

func (s *S3Storage) Size(name string) (int64, error) {
	return 0, errors.New("S3Storage is not implemented.")
}

func (s *S3Storage) ListDir(path string) ([]string, []string, error) {
	return nil, nil, errors.New("S3Storage is not implemented.")
}

func (s *S3Storage) GetAvailableName(name string) string {
	return name
}
