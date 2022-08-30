package s3file

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Clienter interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
}

// S3Object implements the ReaderAt and ReadSeeker interfaces.  This makes S3 files useful in
// applications that expect file like objects.
type S3File struct {
	Bucket string
	Key    string

	s3Client S3Clienter
	offset   int // Position in file
	info     *S3FileInfo
}

// NewS3Object initialises the S3Object.
func NewS3File(bucket, key string, s3c S3Clienter) (*S3File, error) {
	o := S3File{
		Bucket:   bucket,
		Key:      key,
		s3Client: s3c,
	}

	var err error
	o.info, err = o.Stat()
	return &o, err
}

func (o *S3File) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// ReadAt reads bytes into p starting at byte offset off from a file stored in AWS S3. N.B every
// HTTP request will return EOF but io.SectionReader.ReadAt controls when EOF should be returned. In
// summary, never return EOF, this method never knows the big picture of how each request is being
// used in the background. Ref io/io.go:538 SectionReader.ReadAt(..)
func (o *S3File) ReadAt(p []byte, off int64) (n int, err error) {
	// Default read from offset to end of file
	var byteRange = fmt.Sprintf("bytes=%d-", off)
	var size = int64(len(p))

	// Limit the byte range to the buffer length. N.B this will null pointer error if S3Object is
	// not created with NewS3Object. o.info.size is the size of the entire zip archive not the
	// 'file' within it that we are reading.
	if finalByte := off + size - 1; finalByte < o.info.size {
		byteRange += fmt.Sprintf("%d", finalByte)
	}

	res, err := o.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &o.Bucket,
		Key:    &o.Key,
		Range:  &byteRange,
	})
	if err != nil {
		return 0, err
	}

	n, err = res.Body.Read(p)
	// EOF is returned on byte range requests. But we do not know when the overarching Reader is
	// done. It will give us the buffer p to fill.
	if err == io.EOF {
		err = nil
	}

	return n, err
}

func (o *S3File) Stat() (*S3FileInfo, error) {
	var err error
	if o.info == nil {
		o.info = &S3FileInfo{}

		var res *s3.HeadObjectOutput
		if res, err = o.s3Client.HeadObject(&s3.HeadObjectInput{
			Bucket: &o.Bucket,
			Key:    &o.Key,
		}); err == nil {
			o.info.name = fmt.Sprintf("s3://%s/%s", o.Bucket, o.Key)
			o.info.size = *res.ContentLength
			o.info.modtime = *res.LastModified
		}
	}

	return o.info, err
}
