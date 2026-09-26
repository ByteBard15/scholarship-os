package filestore

import (
	"context"
	"io"
)

type SavedFile struct {
	Key    string
	Size   int64
	SHA256 string
}

type FileStore interface {
	Save(context.Context, string, string, io.Reader) (SavedFile, error)
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
