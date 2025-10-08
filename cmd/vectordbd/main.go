package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"vectraDB/pkg/api"
	"vectraDB/pkg/store"
)

var version = "v0.1.0"

func main() {
	vs := store.NewVectorStore()
	apiServer := api.NewAPI(vs)

	r := chi.NewRouter()
	r.Mount("/", apiServer.Routes())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Vectra DB running on port %s\n", port)
	http.ListenAndServe(":"+port,r)
}
