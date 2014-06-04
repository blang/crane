package main

import (
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
	metaStorage   MetaStorage
	fileStorage   FileStorage
	authenticator Authenticator
}

func NewRegistry(metaStorage MetaStorage, fileStorage FileStorage, authenticator Authenticator) *Registry {
	return &Registry{
		metaStorage:   metaStorage,
		fileStorage:   fileStorage,
		authenticator: authenticator,
	}
}

func (r *Registry) SetImageJSON(imageID string, json string) error {
	return r.metaStorage.SetImageJSON(imageID, json)
}
func (r *Registry) ImageJSON(imageID string) (string, bool) {
	return r.metaStorage.ImageJSON(imageID)
}
func (r *Registry) SetChecksum(imageID string, checksum string) error {
	return r.metaStorage.SetChecksum(imageID, checksum)
}
func (r *Registry) Checksum(imageID string) (string, bool) {
	return r.metaStorage.Checksum(imageID)
}

func (r *Registry) Size(imageID string) (int64, bool) {
	return r.metaStorage.Size(imageID)
}

func (r *Registry) Layer(imageID string) (ReadCloseSeeker, error) {
	return r.fileStorage.Layer(imageID)
}
func (r *Registry) SetLayer(imageID string, imageJSON string, reader io.ReadCloser) error {
	checksum, size, err := r.fileStorage.SetLayer(imageID, imageJSON, reader)
	if err == nil {
		//TODO: Check for errors
		log.Printf("Put Layer of image %s with checksum %s", imageJSON, checksum)
		r.metaStorage.SetChecksum(imageID, checksum)
		r.metaStorage.SetSize(imageID, size)
	}
	return err
}

func (r *Registry) Tag(namespace string, repository string, tag string) (string, bool) {
	return r.metaStorage.Tag(namespace, repository, tag)
}

func (r *Registry) Tags(namespace string, repository string) (map[string]string, bool) {
	return r.metaStorage.Tags(namespace, repository)
}
func (r *Registry) SetTag(namespace string, repository string, imageID string, tag string) error {
	return r.metaStorage.SetTag(namespace, repository, imageID, tag)
}

func (r *Registry) SetImages(namespace string, repository string, images []string) error {
	return r.metaStorage.SetImages(namespace, repository, images)
}
func (r *Registry) Images(namespace string, repository string) ([]string, error) {
	return r.metaStorage.Images(namespace, repository)
}
func (r *Registry) Ancestry(imageID string) ([]string, error) {
	return r.metaStorage.Ancestry(imageID)
}

func (r *Registry) SetAncestry(imageID string, parentImageID string) error {
	return r.metaStorage.SetAncestry(imageID, parentImageID)
}

func (r *Registry) Authenticator() Authenticator {
	return r.authenticator
}
