package main

import (
	"flag"
	"github.com/blang/crane/auth"
	"github.com/blang/crane/store"
	"log"
	"net/http"
)

func main() {
	listen := flag.String("listen", ":5000", "Addr to listen on")
	dataDir := flag.String("datadir", "/tmp/registry", "Data directory")
	flag.Parse()

	metaStorage := store.NewMemMetaStorage()
	fileStorage := store.NewLocalFileStorage(*dataDir)
	authenticator := auth.NewLocalAuthenticator()
	proxyStore := store.NewProxyStore(metaStorage, fileStorage)
	registry := NewRegistry(proxyStore, authenticator)
	api := NewRegistryAPI(registry)

	log.Printf("Starting server listening on %q", *listen)
	if err := http.ListenAndServe(*listen, api); err != nil {
		log.Fatalf("HTTP Server crashed: %v", err)
	}
}
