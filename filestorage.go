package main

import (
	"io"
)

// Interface for file storage like layers
type FileStorage interface {
	Layer(string) (io.Reader, error)
	SetLayer(string, io.Reader) error
}
