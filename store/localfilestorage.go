package store

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path"
)

type LocalFileStorage struct {
	dataDir string
}

func NewLocalFileStorage(dataDir string) *LocalFileStorage {
	err := os.MkdirAll(dataDir, 0755)
	log.Printf("Directory created %s", dataDir)
	if err != nil {
		panic("Could not create local file storage:" + err.Error())
	}
	return &LocalFileStorage{
		dataDir: dataDir,
	}
}

func (s *LocalFileStorage) Layer(imageID string) (ReadCloseSeeker, error) {
	layerPath := path.Join(s.dataDir, imageID+"_layer")
	f, err := os.OpenFile(layerPath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *LocalFileStorage) SetTmpLayer(imageID string, imageJSON string, r io.ReadCloser) (string, int64, error) {
	layerTmpPath := path.Join(s.dataDir, imageID+"_layer.tmp")
	w, err := os.OpenFile(layerTmpPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return "", 0, err
	}
	defer w.Close()
	defer r.Close()
	//Hashing
	hash := sha256.New()
	hash.Write([]byte(imageJSON + "\n"))
	// make a buffer to keep chunks that are read
	buf := make([]byte, 4096)
	var read int64 = 0
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return "", 0, err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return "", 0, err
		}
		if _, err := hash.Write(buf[:n]); err != nil {
			return "", 0, err
		}
		read += int64(n)
	}

	hashBytes := hash.Sum(nil)
	mdStr := hex.EncodeToString(hashBytes)
	return mdStr, read, nil
}

func (s *LocalFileStorage) CommitTmpLayer(imageID string) bool {
	layerTmpPath := path.Join(s.dataDir, imageID+"_layer.tmp")
	layerPath := path.Join(s.dataDir, imageID+"_layer")
	r, err := os.OpenFile(layerTmpPath, os.O_RDONLY, 0600)
	if err != nil {
		return false
	}
	r.Close()
	err = os.Rename(layerTmpPath, layerPath)
	if err != nil {
		return false
	}
	return true
}
func (s *LocalFileStorage) DiscardTmpLayer(imageID string) bool {
	layerTmpPath := path.Join(s.dataDir, imageID+"_layer.tmp")
	err := os.Remove(layerTmpPath)
	if err != nil {
		return false
	}
	return true
}
