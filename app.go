package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	listen := flag.String("listen", ":5000", "Addr to listen on")
	dataDir := flag.String("datadir", "/tmp/registry", "Data directory")
	flag.Parse()

	registry := NewRegistry(*dataDir)
	api := NewRegistryAPI(registry)

	log.Printf("Starting server listening on %q", *listen)
	if err := http.ListenAndServe(*listen, api); err != nil {
		log.Fatalf("HTTP Server crashed: %v", err)
	}
}
