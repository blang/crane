package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blang/crane/auth"
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
	r.router.HandleFunc("/v1/users/", r.handlePostUser).Methods("POST")
	r.router.HandleFunc("/v1/users/", r.handleGetUser).Methods("GET")
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

// Handles User Registration: Returns 400 to redirect docker to GET /v1/users
// Route: POST /v1/users
func (r *RegistryAPI) handlePostUser(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("\"Username or email already exists\""))
	//TODO: Maybe give the authenticator a try on this
	r.handleDummy(w, req)
}

// Handles User Login: Returns 200 because Auth is handled by proxy
// Status: 200 : Login successful
// Status: 401 : Wrong login
// Status: 403 : Account inactive
// Route: GET /v1/users
func (r *RegistryAPI) handleGetUser(w http.ResponseWriter, req *http.Request) {
	user, pass, valid := authHeader(req)
	if !valid {
		log.Println("Authheader not valid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().Authenticate(user, pass) {
		log.Println("Authentication failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
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
	log.Println("Put Repository")
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		log.Println("Put Repository namespace, repository url wrong")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, pass, valid := authHeader(req)
	if !valid {
		log.Printf("Not valid header")
		w.WriteHeader(http.StatusUnauthorized)
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

	token, granted := r.registry.Authenticator().Authorize(user, pass, namespace, repository, imageIds, auth.O_WRONLY)
	if !granted {
		log.Printf("Not granted %s", repository)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	setTokenHeaders(w, token, namespace, repository, auth.O_WRONLY)

	err = r.registry.SetImages(namespace, repository, imageIds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("X-Docker-Endpoints", req.Host) //TODO: Make http host or config
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
	w.Header().Set("X-Docker-Payload-Checksum", "sha256:"+checksum)
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
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().HasPermPushImage(token, imageID) {
		w.WriteHeader(http.StatusUnauthorized)
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
	err = r.registry.SetTmpImageJSON(imageID, string(bImageJSON))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if newImage.Parent != "" {
		r.registry.SetTmpAncestry(imageID, newImage.Parent)
	}
}

func (r *RegistryAPI) handlePutImageChecksum(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	imageID, found := vars["image_id"]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().HasPermPushImage(token, imageID) {
		w.WriteHeader(http.StatusUnauthorized)
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

	if !r.registry.ValidateAndCommitLayer(imageID, checksum) {
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
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().HasPermPushImage(token, imageID) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	imageJSON, found := r.registry.TmpImageJSON(imageID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := r.registry.SetTmpLayer(imageID, imageJSON, req.Body)
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
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().HasPermPullImage(token, imageID) {
		w.WriteHeader(http.StatusUnauthorized)
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
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	imageID := strings.Trim(string(bImageID), "\" ")
	if !r.registry.Authenticator().HasPermPushTag(token, namespace, repository, imageID, tags) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !r.registry.Authenticator().HasPermPullTag(token, namespace, repository, tag) {
		w.WriteHeader(http.StatusUnauthorized)
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

	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !r.registry.Authenticator().HasPermPullTags(token, namespace, repository) {
		w.WriteHeader(http.StatusUnauthorized)
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

	// No Auth here because it's an empty endpoint result of dockers strange design.

	b, _ := ioutil.ReadAll(req.Body)
	bodyStr := string(b)
	log.Printf("Put Repository Images: %s:%s %v", namespace, repository, string(b))
	if bodyStr == "[]" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusBadRequest)

}

func (r *RegistryAPI) handleGetRepositoryImages(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespace, found1 := vars["namespace"]
	repository, found2 := vars["repository"]
	if !(found1 && found2) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, pass, valid := authHeader(req)
	if !valid {
		log.Printf("Not valid header")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	images, err := r.registry.Images(namespace, repository)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		//TODO: Maybe security problem because guessing is possible
		return
	}

	token, granted := r.registry.Authenticator().Authorize(user, pass, namespace, repository, images, auth.O_RDONLY)
	if !granted {
		log.Printf("Not granted %s", repository)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	setTokenHeaders(w, token, namespace, repository, auth.O_RDONLY)

	w.Header().Set("X-Docker-Endpoints", req.Host) //TODO: Make http host or config

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
	token, validToken := tokenHeader(req)
	if !validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.registry.Authenticator().HasPermPullImage(token, imageID) {
		w.WriteHeader(http.StatusUnauthorized)
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

func authHeader(req *http.Request) (string, string, bool) {
	const authBasic = "Basic "
	auth := req.Header.Get("Authorization")
	log.Printf("Header is %s", auth)
	if !strings.HasPrefix(auth, authBasic) {
		log.Println("Has no prefx")
		return "", "", false
	}
	str, err := base64.StdEncoding.DecodeString(auth[len(authBasic):])
	if err != nil {
		return "", "", false
	}
	log.Printf("Auth string is: %s", str)

	creds := strings.SplitN(string(str), ":", 2)
	if len(creds) != 2 {
		return "", "", false
	}
	return creds[0], creds[1], true
}

func tokenHeader(req *http.Request) (string, bool) {
	const PREFIX = "Token Token "
	const SIGPREFIX = "signature="
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	if !strings.HasPrefix(auth, PREFIX) {
		return "", false
	}

	auth = strings.TrimPrefix(auth, PREFIX)
	parts := strings.Split(auth, ",")
	if len(parts) != 3 {
		return "", false
	}

	signature := ""
	for _, part := range parts {
		if strings.HasPrefix(part, SIGPREFIX) {
			signature = part[len(SIGPREFIX):]
		}
	}
	if signature == "" {
		return "", false
	}

	return signature, true
}

func setTokenHeaders(w http.ResponseWriter, token string, namespace string, repository string, mode auth.Mode) {
	modeStr := ""
	switch mode {
	case auth.O_RDONLY:
		modeStr = "read"
	case auth.O_WRONLY:
		modeStr = "write"
	default:
		modeStr = "read"
	}

	tokenHeader := fmt.Sprintf("Token signature=%s,repository=\"%s/%s\",access=%s", token, namespace, repository, modeStr)
	w.Header().Set("WWW-Authenticate", tokenHeader)
	w.Header().Set("X-Docker-Token", tokenHeader)
}
