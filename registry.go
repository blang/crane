package main

import (
	"github.com/blang/crane/auth"
	"github.com/blang/crane/store"
	"io"
	"log"
)

type Image struct {
	ID              string `json:"id"`
	Parent          string `json:"parent,omitempty"`
	Comment         string `json:"comment"`
	Created         string `json:"created"`
	ContainerConfig ContainerConfig
	DockerVersion   string `json:"docker_version"`
	Architecture    string `json:"architecture"`
}

type ContainerConfig struct {
	Hostname     string `json:"Hostname"`
	User         string `json:"User"`
	Memory       string `json:"Memory"`
	MemorySwap   string `json:"MemorySwap"`
	CPUShares    string `json:"CpuShares"`
	AttachStdin  bool   `json:"AttachStdin"`
	AttachStdout bool   `json:"AttachStdout"`
	AttachStderr bool   `json:"AttachStderr"`
	PortSpecs    bool   `json:"PortSpecs"`
	TTY          bool   `json:"Tty"`
	OpenStdin    bool   `json:"OpenStdin"`
	StdinOnce    bool   `json:"StdinOnce"`
	Env          string `json:"Env"`
	Cmd          string `json:"Cmd"`
	Dns          string `json:"Dns"`
	Image        string `json:"Image"`
	Volumes      string `json:"Volumes"`
	VolumesFrom  string `json:"VolumesFrom"`
}

type Registry struct {
	store         store.Store
	authenticator auth.Authenticator
}

func NewRegistry(store store.Store, authenticator auth.Authenticator) *Registry {
	return &Registry{
		store:         store,
		authenticator: authenticator,
	}
}

func (r *Registry) SetTmpImageJSON(imageID string, json string) error {
	return r.store.SetTmpImageJSON(imageID, json)
}

func (r *Registry) ImageJSON(imageID string) (string, bool) {
	return r.store.ImageJSON(imageID)
}

func (r *Registry) TmpImageJSON(imageID string) (string, bool) {
	return r.store.TmpImageJSON(imageID)
}

func (r *Registry) Checksum(imageID string) (string, bool) {
	return r.store.Checksum(imageID)
}

func (r *Registry) Size(imageID string) (int64, bool) {
	return r.store.Size(imageID)
}

func (r *Registry) Layer(imageID string) (store.ReadCloseSeeker, error) {
	return r.store.Layer(imageID)
}

func (r *Registry) SetTmpLayer(imageID string, imageJSON string, reader io.ReadCloser) error {
	checksum, size, err := r.store.SetTmpLayer(imageID, imageJSON, reader)
	if err == nil {
		//TODO: Check for errors
		log.Printf("Put Tmp Layer of image %s with checksum %s", imageJSON, checksum)
		r.store.SetTmpChecksum(imageID, checksum)
		r.store.SetTmpSize(imageID, size)
	}
	return err
}

func (r *Registry) Tag(namespace string, repository string, tag string) (string, bool) {
	return r.store.Tag(namespace, repository, tag)
}

func (r *Registry) Tags(namespace string, repository string) (map[string]string, bool) {
	return r.store.Tags(namespace, repository)
}

func (r *Registry) SetTag(namespace string, repository string, imageID string, tag string) error {
	return r.store.SetTag(namespace, repository, imageID, tag)
}

func (r *Registry) SetImages(namespace string, repository string, images []string) error {
	return r.store.SetImages(namespace, repository, images)
}

func (r *Registry) Images(namespace string, repository string) ([]string, error) {
	return r.store.Images(namespace, repository)
}

func (r *Registry) Ancestry(imageID string) ([]string, error) {
	return r.store.Ancestry(imageID)
}

func (r *Registry) SetTmpAncestry(imageID string, parentImageID string) error {
	return r.store.SetTmpAncestry(imageID, parentImageID)
}

func (r *Registry) Authenticator() auth.Authenticator {
	return r.authenticator
}

func (r *Registry) ValidateAndCommitLayer(imageID string, checksum string) bool {
	tmpChs, found := r.store.TmpChecksum(imageID)
	if !found {
		r.discardImage(imageID)
		return false
	}
	if tmpChs != checksum {
		r.discardImage(imageID)
		return false
	}
	succ := r.store.CommitTmpLayer(imageID)
	if !succ {
		r.discardImage(imageID)
		return false
	}
	succ = r.store.CommitTmpImage(imageID)
	if !succ {
		r.discardImage(imageID)
		return false
	}
	return true

}

func (r *Registry) discardImage(imageID string) bool {
	r1 := r.store.DiscardTmpImage(imageID)
	r2 := r.store.DiscardTmpLayer(imageID)
	return (r1 && r2)
}
