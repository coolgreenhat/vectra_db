package main

import (
	"fmt"
	"net/http"
	"os"
	"log"

	"github.com/go-chi/chi/v5"
	"vectraDB/pkg/api"
	"vectraDB/pkg/store"
)

var version = "v0.1.0"

func main() {
	dbPath := "vectra.db"
	vs, err := store.NewVectorStore(dbPath)
	
	if err != nil {
		log.Fatalf("Failed to initialize vector store: %v")
	}

	apiServer := api.NewAPI(vs)

	r := chi.NewRouter()
	r.Mount("/", apiServer.Routes())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Vectra DB running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port,r))
}
