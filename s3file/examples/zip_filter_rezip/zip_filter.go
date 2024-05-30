package main

import (
	"archive/zip"
	"fmt"
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

func zipReadWrite(f *zip.File, outW *zip.Writer) error {
	// Read file data from inside zip
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("error opening zip subfile: %w", err)
	}
	defer rc.Close()

	// Consume file data. N.B File must be completely read before the next call to `Create`
	w, err := outW.Create(f.Name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, rc)
	return err
}

// extractJSON does not close anything. If you need it closed add it to `closer`.
func extractJSON(r *zip.Reader, outW *zip.Writer, closer func()) chan error {
	errC := make(chan error, 1)

	// Iterate through the files in the archive, looking for json files.
	go func() {
		defer closer()
		for _, f := range r.File {
			log.Println(f.Name)

			if !strings.HasSuffix(f.Name, ".json") {
				log.Println("not found json file. Skipping ...")
				continue
			}

			if err := zipReadWrite(f, outW); err != nil {
				errC <- err
				return
			}

			log.Printf("Successfully wrote file '%s'\n", f.Name)
		}
		errC <- nil
	}()
	return errC
}

/*
gac -p rfs-dev

// Prepate and upload test archive
ls -l ../archive/
zip -r --junk-paths archive.zip ../archive/
unzip -l archive.zip
aws s3 cp archive.zip s3://telson-temp-bucket/

// Remove existing result
aws s3 rm s3://telson-temp-bucket/multi-json.zip

aws s3 ls s3://telson-temp-bucket/

go run zip_filter.go

// Verify results
aws s3 cp s3://telson-temp-bucket/multi-json.zip /tmp

unzip -l /tmp/multi-json.zip
unzip /tmp/multi-json.zip -d /tmp/multi-temp/

*/
func main() {
	f, err := s3file.NewS3File("telson-temp-bucket", "archive.zip", s3Client)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	// Open source
	inR, err := zip.NewReader(f, fi.Size())
	if err != nil {
		log.Fatal(err)
	}

	// Open destination
	outR, w := io.Pipe()
	outW := zip.NewWriter(w)

	// Start reading/writting
	errC := extractJSON(inR, outW, func() {
		outW.Close()
		w.Close()
	})

	// Consume writter to S3
	bucket, key := "telson-temp-bucket", "multi-json.zip"
	log.Printf("Starting PutStream to s3://%s/%s ... \n", bucket, key)
	if err := PutStream(outR, bucket, key); err != nil {
		log.Fatalf("error sending file to s3: %v", err)
	}

	err = <-errC
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DONE")
}
