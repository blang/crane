package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

//Image json
// {
//     "id": "511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158",
//     "comment": "Imported from -",
//     "created": "2013-06-13T14:03:50.821769-07:00",
//     "container_config": {
//         "Hostname": "",
//         "User": "",
//         "Memory": 0,
//         "MemorySwap": 0,
//         "CpuShares": 0,
//         "AttachStdin": false,
//         "AttachStdout": false,
//         "AttachStderr": false,
//         "PortSpecs": null,
//         "Tty": false,
//         "OpenStdin": false,
//         "StdinOnce": false,
//         "Env": null,
//         "Cmd": null,
//         "Dns": null,
//         "Image": "",
//         "Volumes": null,
//         "VolumesFrom": ""
//     },
//     "docker_version": "0.4.0",
//     "architecture": "x86_64"
// }

var (
	JsonMsgImageNotFound = []byte("{\"error\": \"Image not found\"}")
	JsonMsgPing          = []byte("true")
)

type RegistryAPI struct {
	router   *mux.Router
	registry *Registry
}

func NewRegistryAPI(registry *Registry) *RegistryAPI {
	r := &RegistryAPI{
		router:   mux.NewRouter(),
		registry: registry,
	}
	r.registerEndPoints()
	return r
}

func (r *RegistryAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Docker-RegistryAPI-Version", "0.1.0")
	log.Printf("%s: %s", req.Method, req.RequestURI)
	r.router.ServeHTTP(w, req)
}

func (r *RegistryAPI) registerEndPoints() {

	r.router.HandleFunc("/v1/_ping", r.handlePing).Methods("GET")

	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/", r.handlePutRepository).Methods("PUT")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}", r.handleDummy).Methods("DELETE")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags", r.handleDummy).Methods("GET")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handleDummy).Methods("GET")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handleDummy).Methods("PUT")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handleDummy).Methods("DELETE")

	r.router.HandleFunc("/v1/images/{image_id}/json", r.handleGetImageJson).Methods("GET")
	r.router.HandleFunc("/v1/images/{image_id}/json", r.handlePutImageJson).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/layer", r.handleDummy).Methods("GET")
	r.router.HandleFunc("/v1/images/{image_id}/layer", r.handleDummy).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/checksum", r.handleDummy).Methods("GET")
	r.router.HandleFunc("/v1/images/{image_id}/checksum", r.handleDummy).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/ancestry", r.handleDummy).Methods("GET")
	//
}

func (r *RegistryAPI) handleDummy(w http.ResponseWriter, req *http.Request) {
	b, _ := ioutil.ReadAll(req.Body)
	if len(b) < 500 {
		log.Printf("Body: %s (%d)", b, len(b))
	} else {
		log.Printf("Body length: %s", len(b))
	}
	log.Printf("Header: %s", req.Header)
}

type RepositoryPut struct {
	ID  string  `json:"id"`
	Tag *string `json:"Tag"`
}

func (rp RepositoryPut) String() string {
	return "id:" + rp.ID
}

func (r *RegistryAPI) handlePutRepository(w http.ResponseWriter, req *http.Request) {
	var images []RepositoryPut
	err := json.NewDecoder(req.Body).Decode(&images)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("WWW-Authenticate", "Token signature=HN2IR0O98XB1GEJA,repository=\"test/busybox\",access=write")
	w.Header().Set("X-Docker-Endpoints", "127.0.0.1:5001")
	w.Header().Set("X-Docker-Token", "Token signature=HN2IR0O98XB1GEJA,repository=\"test/busybox\",access=write")
	log.Printf("Images: %q", images)
}

func (r *RegistryAPI) handleGetImageJson(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	image_id, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	image := r.registry.ImageByID(image_id)
	if image == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgImageNotFound)
		return
	}

	b, err := json.Marshal(image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (r *RegistryAPI) handlePutImageJson(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	image_id, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	image := r.registry.ImageByID(image_id)
	if image != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	var newImage Image
	err := json.NewDecoder(req.Body).Decode(&newImage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.registry.AddImageJSON(image_id, &newImage)
}

func (r *RegistryAPI) handlePing(w http.ResponseWriter, req *http.Request) {
	w.Write(JsonMsgPing)
}
