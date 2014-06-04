package main

import (
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
	authenticator Authenticator
}

func NewRegistry(store store.Store, authenticator Authenticator) *Registry {
	return &Registry{
		store:         store,
		authenticator: authenticator,
	}
}

func (r *Registry) SetImageJSON(imageID string, json string) error {
	return r.store.SetImageJSON(imageID, json)
}
func (r *Registry) ImageJSON(imageID string) (string, bool) {
	return r.store.ImageJSON(imageID)
}
func (r *Registry) SetChecksum(imageID string, checksum string) error {
	return r.store.SetChecksum(imageID, checksum)
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
func (r *Registry) SetLayer(imageID string, imageJSON string, reader io.ReadCloser) error {
	checksum, size, err := r.store.SetLayer(imageID, imageJSON, reader)
	if err == nil {
		//TODO: Check for errors
		log.Printf("Put Layer of image %s with checksum %s", imageJSON, checksum)
		r.store.SetChecksum(imageID, checksum)
		r.store.SetSize(imageID, size)
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

func (r *Registry) SetAncestry(imageID string, parentImageID string) error {
	return r.store.SetAncestry(imageID, parentImageID)
}

func (r *Registry) Authenticator() Authenticator {
	return r.authenticator
}
