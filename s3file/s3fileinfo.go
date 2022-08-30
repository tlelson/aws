package s3file

import (
	"io/fs"
	"time"
)

// S3ObjectInfo implements fs.FileInfo
type S3FileInfo struct {
	name    string
	size    int64
	modtime time.Time
}

func (oi S3FileInfo) Name() string       { return oi.name }
func (oi S3FileInfo) Size() int64        { return oi.size }
func (oi S3FileInfo) Mode() fs.FileMode  { return fs.ModeIrregular }
func (oi S3FileInfo) ModTime() time.Time { return oi.modtime }
func (oi S3FileInfo) IsDir() bool        { return false } // Can't HeadObject non objects
func (oi S3FileInfo) Sys() any           { return "s3" }
