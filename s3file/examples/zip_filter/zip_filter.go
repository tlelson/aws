package main

import (
	"archive/zip"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/tlelson/aws/s3file"
)

var s3Client *s3.S3 = s3.New(session.Must(session.NewSession(&aws.Config{
	Region: aws.String("ap-southeast-2"),
})))

func PutStream(data io.Reader, bucket, key string) error {
	uploader := s3manager.NewUploaderWithClient(s3Client)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	return err
}

// writeFile writes the contents of f to writter w.  It may be called in a seperate go routine.
func writeFile(f *zip.File, w io.WriteCloser) {
	log.Println("Starting writter ... ")

	rc, err := f.Open()
	if err != nil {
		log.Fatalf("error opening zip subfile: %s", err)
	}
	defer rc.Close()
	written, err := io.Copy(w, rc)
	w.Close()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Copy complete. %v bytes written to writter", written)
}

func main() {
	// Open zip archive. Similar to:
	// f, err := os.Open("./archive.zip")
	f, err := s3file.NewS3File("SRC_BUCKET", "archive.zip", s3Client)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	r, err := zip.NewReader(f, fi.Size())
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the files in the archive, looking for json files.
	for _, f := range r.File {
		log.Println(f.Name)

		if !strings.HasSuffix(f.Name, ".json") {
			log.Println("not found json file. Skipping ...")
			continue
		}

		// Get Writer for output. Similar to file writer (except that the following would not use
		// async `writeFile` because the file is already open):
		// w, err := os.OpenFile(wfPath, os.O_RDWR|os.O_CREATE, 0644)
		data, w := io.Pipe()
		go writeFile(f, w)

		bucket := "DST_BUCKET"
		log.Printf("Starting PutStream to s3://%s/%s ... \n", bucket, f.Name)
		if err := PutStream(data, bucket, f.Name); err != nil {
			log.Fatalf("error sending file to s3: %v", err)
		}

		log.Printf("Successfully wrote file\n")
	}
	log.Println("DONE")
}
