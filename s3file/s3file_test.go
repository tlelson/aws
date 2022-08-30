package s3file_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tlelson/aws/s3file"
)

type mockS3Client struct {
	headObjectOutput *s3.HeadObjectOutput
}

func (s3c mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{}, nil //fmt.Errorf("not implemented")
}
func (s3c mockS3Client) HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return s3c.headObjectOutput, nil
}

func ptr[T any](x T) *T { return &x }

func TestNewS3File(t *testing.T) {

	s3c := mockS3Client{&s3.HeadObjectOutput{
		ContentLength: ptr(int64(15)),
		LastModified:  ptr(time.Now()),
	}}

	_, err := s3file.NewS3File("bucket", "key", s3c)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
