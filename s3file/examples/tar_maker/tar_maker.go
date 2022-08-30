/*
Demonstrates how S3 files can be tar-ed up in memory and written back to S3.

based on example: https://pkg.go.dev/archive/tar#example-package-Minimal
*/

package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/tlelson/aws/s3file"
)

var s3Client *s3.S3 = s3.New(session.Must(session.NewSession(&aws.Config{
	Region: aws.String("ap-southeast-2"),
})))

// Create an uploader with the session and default options
func PutStream(data io.Reader, bucket, key string) error {
	uploader := s3manager.NewUploaderWithClient(s3Client)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	return err
}

type s3Object struct {
	Bucket, Key string
}

// tarS3objects writes a tar archive to	`w` as it reads and tars up s3objects. Writes block until a
// corresponding Read operation takes place.
func tarS3objects(s3objects []s3Object, w io.WriteCloser) {
	tw := tar.NewWriter(w)
	defer func() {
		// Finished the tar file footer
		if err := tw.Close(); err != nil {
			log.Fatal(err)
		}
		// Closes the output stream
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for _, file := range s3objects {
		log.Printf("Loading file s3://%s/%s\n", file.Bucket, file.Key)
		s3f, err := s3file.NewS3File(file.Bucket, file.Key, s3Client)
		if err != nil {
			log.Fatalln(err)
		}
		s3finfo, err := s3f.Stat()
		if err != nil {
			log.Fatalln(err)
		}

		hdr := &tar.Header{
			Name: file.Key,
			Mode: 0600,
			Size: s3finfo.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln("Error writing header:", err)
		}

		n, err := io.Copy(tw, s3f)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Successfully wrote %d bytes of %s\n", n, s3finfo.Name())
	}
}

func main() {
	// Files to archive
	var files = []s3Object{
		{"BUCKET", "archive/one.txt"},
		{"BUCKET", "archive/two.txt"},
	}

	// Write to S3
	output, w := io.Pipe()
	go tarS3objects(files, w)

	bucket, key := "BUCKET", "tar_archive.tar"
	log.Printf("Starting PutStream to s3://%s/%s ... \n", bucket, key)
	err := PutStream(output, bucket, key)
	log.Println("Finished PutStream with error: ", err)

	fmt.Println("DONE")
}
