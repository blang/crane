package store

import (
	"io"
)

type ReadCloseSeeker interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Interface for file storage like layers
type FileStorage interface {
	Layer(imageID string) (ReadCloseSeeker, error)
	SetLayer(imageID string, imageJSON string, r io.ReadCloser) (string, int64, error) // Closes r
}
