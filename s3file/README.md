# S3File 

Implements `Read`, `Seek`, `ReadAt` methods on S3 files so that operations typically expecting 'file-like' objects are able to use S3 files.


## Example 1 - JSON filter

Suppose you have large zip archives in S3.  They include large video files, but
you only want the json metadata file stored with it. You want to do it in
lambda without block storage (in memory).

[JSON Filter](./examples/zip_filter/zip_filter.go)
