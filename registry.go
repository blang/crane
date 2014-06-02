package main

import ()

type Image struct {
	ID              string `json:"id"`
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
	dataDir  string
	imageMap map[string]*Image //TODO: Not Threadsafe
}

func NewRegistry(dataDir string) *Registry {
	return &Registry{
		dataDir:  dataDir,
		imageMap: make(map[string]*Image),
	}
}

func (r *Registry) ImageByID(id string) *Image {
	if image, found := r.imageMap[id]; found {
		return image
	}
	return nil
}

func (r *Registry) AddImageJSON(id string, image *Image) {
	r.imageMap[id] = image
}
