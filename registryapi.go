package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	JsonMsgImageNotFound         = []byte("{\"error\": \"Image not found\"}")
	JsonMsgTagNotFound           = []byte("{\"error\": \"Tag not found\"}")
	JsonMsgTagsNotFound          = []byte("{\"error\": \"No tags found\"}")
	JsonMsgPing                  = []byte("true")
	JsonMsgImageMissingChecksum  = []byte("{\"error\": \"Cannot set this image checksum\"}")
	JsonMsgImageChecksumMissing  = []byte("{\"error\": \"Missing Image's checksum\"}")
	JsonMsgImageChecksumMismatch = []byte("{\"error\": \"Checksum mismatch\"}")
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
	log.Printf("Header: %s", req.Header)
	log.Printf("Content-Length: %d", req.ContentLength)
	r.router.ServeHTTP(w, req)
}

func (r *RegistryAPI) registerEndPoints() {

	r.router.HandleFunc("/v1/_ping", r.handlePing).Methods("GET")

	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/", r.handlePutRepository).Methods("PUT")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/images", r.handlePutRepositoryImages).Methods("PUT")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/images", r.handleGetRepositoryImages).Methods("GET")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}", r.handleDummy).Methods("DELETE")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags", r.handleGetRepositoryTags).Methods("GET")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handleGetRepositoryTag).Methods("GET")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handlePutRepositoryTag).Methods("PUT")
	r.router.HandleFunc("/v1/repositories/{namespace}/{repository}/tags/{tags}", r.handleDummy).Methods("DELETE")

	r.router.HandleFunc("/v1/images/{image_id}/json", r.handleGetImageJson).Methods("GET")
	r.router.HandleFunc("/v1/images/{image_id}/json", r.handlePutImageJson).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/layer", r.handleGetImageLayer).Methods("GET")
	r.router.HandleFunc("/v1/images/{image_id}/layer", r.handlePutImageLayer).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/checksum", r.handlePutImageChecksum).Methods("PUT")

	r.router.HandleFunc("/v1/images/{image_id}/ancestry", r.handleGetAncestry).Methods("GET")
	//
}

func (r *RegistryAPI) handleDummy(w http.ResponseWriter, req *http.Request) {
	b, _ := ioutil.ReadAll(req.Body)
	if len(b) < 500 {
		log.Printf("Body: %s (%d)", b, len(b))
	} else {
		log.Printf("Body length: %s", len(b))
	}

}

type ImageRef struct {
	ID  string  `json:"id"`
	Tag *string `json:"Tag,omitempty"`
}

func (r ImageRef) String() string {
	return "id:" + r.ID
}

func (r *RegistryAPI) handlePutRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var imageRefs []ImageRef
	err := json.NewDecoder(req.Body).Decode(&imageRefs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var imageIds []string
	for _, imageRef := range imageRefs {
		imageIds = append(imageIds, imageRef.ID)
	}
	err = r.registry.SetImages(namespace, repository, imageIds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("WWW-Authenticate", "Token signature=HN2IR0O98XB1GEJA,repository=\"test/busybox\",access=write")
	w.Header().Set("X-Docker-Endpoints", "127.0.0.1:5001")
	w.Header().Set("X-Docker-Token", "Token signature=HN2IR0O98XB1GEJA,repository=\"test/busybox\",access=write")
	log.Printf("Put Repository Images: %s:%s %q", namespace, repository, imageRefs)
}

func (r *RegistryAPI) handleGetImageJson(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageJSON, found := r.registry.ImageJSON(imageID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgImageNotFound)
		return
	}
	checksum, found := r.registry.Checksum(imageID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgImageNotFound)
		return
	}

	size, found := r.registry.Size(imageID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgImageNotFound)
		return
	}
	w.Header().Set("X-Docker-Payload-Checksum", "sha256:"+checksum) //TODO: format: [sha256: ... ]
	w.Header().Set("X-Docker-Size", strconv.FormatInt(size, 10))
	w.Write([]byte(imageJSON))
}

func (r *RegistryAPI) handlePutImageJson(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, found = r.registry.ImageJSON(imageID)
	if found {
		w.WriteHeader(http.StatusConflict)
		// TODO: Possible to overwrite? I guess
		// Then Reset all other data, checksum, size etc
		return
	}

	var newImage Image
	bImageJSON, err := ioutil.ReadAll(req.Body)
	err = json.Unmarshal(bImageJSON, &newImage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = r.registry.SetImageJSON(imageID, string(bImageJSON))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if newImage.Parent != "" {
		r.registry.SetAncestry(imageID, newImage.Parent)
	}
}

func (r *RegistryAPI) handlePutImageChecksum(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	regChecksum, found := r.registry.Checksum(imageID)
	if !found {
		w.WriteHeader(http.StatusConflict)
		w.Write(JsonMsgImageMissingChecksum)
		return
	}
	checksumHeader := req.Header.Get("X-Docker-Checksum-Payload")
	if checksumHeader == "" || len(checksumHeader) < 8 { // Naive check if checksum might be ok
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JsonMsgImageChecksumMissing)
		return
	}
	found256 := strings.Index(checksumHeader, "sha256:")

	if found256 == -1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	startIndex := found256 + len("sha256:")
	checksum := checksumHeader[startIndex:]

	if regChecksum != checksum {
		log.Printf("Checksum wrong: expected %s got %s", regChecksum, checksum)
		w.WriteHeader(http.StatusConflict)
		w.Write(JsonMsgImageChecksumMismatch)
		return
	}
}

func (r *RegistryAPI) handlePutImageLayer(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageJSON, found := r.registry.ImageJSON(imageID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := r.registry.SetLayer(imageID, imageJSON, req.Body)
	if err != nil {
		log.Printf("Could not set layer: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (r *RegistryAPI) handleGetImageLayer(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	reader, err := r.registry.Layer(imageID)
	if err != nil {
		log.Printf("Could not set layer: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.ServeContent(w, req, "layer.bin", time.Now(), reader)

}

func (r *RegistryAPI) handlePing(w http.ResponseWriter, req *http.Request) {
	w.Write(JsonMsgPing)
}

func (r *RegistryAPI) handlePutRepositoryTag(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	tags, found3 := vars["tags"]
	bImageID, err := ioutil.ReadAll(req.Body)
	if !(found1 && found2 && found3 && err == nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageID := strings.Trim(string(bImageID), "\" ")
	log.Printf("Set Tag: %s/%s: %s:%s", namespace, repository, imageID, tags)
	r.registry.SetTag(namespace, repository, imageID, tags)
}

func (r *RegistryAPI) handleGetRepositoryTag(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	tag, found3 := vars["tags"]
	if !(found1 && found2 && found3) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageID, found := r.registry.Tag(namespace, repository, tag)
	if !found {
		log.Printf("Get Tag not found: %s/%s: %s", namespace, repository, tag)
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgTagNotFound)
		return
	}
	log.Printf("Get Tag: %s/%s: %s:%s", namespace, repository, imageID, tag)
	w.Write([]byte("\"" + imageID + "\""))
}

func (r *RegistryAPI) handleGetRepositoryTags(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageIDs, found := r.registry.Tags(namespace, repository)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonMsgTagsNotFound)
		return
	}
	b, err := json.Marshal(&imageIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (r *RegistryAPI) handlePutRepositoryImages(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b, _ := ioutil.ReadAll(req.Body)
	bodyStr := string(b)
	log.Printf("Put Repository Images: %s:%s %v", namespace, repository, string(b))
	if bodyStr == "[]" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

}

func (r *RegistryAPI) handleGetRepositoryImages(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	images, err := r.registry.Images(namespace, repository)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var imageRefList []*ImageRef
	for _, imageID := range images {
		imageRefList = append(imageRefList, &ImageRef{ID: imageID})
	}
	log.Printf("Get Repository Images: %s:%s %v", namespace, repository, images)
	json.NewEncoder(w).Encode(imageRefList)
}

func (r *RegistryAPI) handleGetAncestry(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ancestryArr, err := r.registry.Ancestry(imageID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Printf("Ancestry for image %s: %v", imageID, ancestryArr)
	json.NewEncoder(w).Encode(&ancestryArr)
}
