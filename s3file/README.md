# S3File 

Implements `Read`, `Seek`, `ReadAt` methods on S3 files so that operations
typically expecting 'file-like' objects are able to use S3 files.

## Usage 

1. Ensure valid AWS credentials are present on your system/shell
2. Ensure source files for the given example are present in s3 
3. Change 'bucket' and 'key' fields in examples accordingly
4. Run the code

```
go mod init main 
go mod tidy 
go run example.go
```

## Example 1 -- JSON filter

Suppose you have large zip archives in S3.  They include large video files, but
you only want the json metadata file stored with it. You want to do it in
lambda without block storage (in memory).

[JSON Filter](./examples/zip_filter/zip_filter.go)

## Example 2 -- Tar archive

Compress a list of S3 files into a tar archive on S3 without using disk or
unbounded memory.

[Tar Archive](./examples/tar_maker/tar_maker.go)
