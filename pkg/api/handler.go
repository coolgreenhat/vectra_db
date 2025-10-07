package api

import (
  "encoding/json"
  "net/http"

  "github.com/go-chi/chi/v5"
  "github.com/coolgreenhat/vectra_db/pkg/store"
)

type API struct {
	Store *store.VectorStore
}

func NewAPI(store *store.VectorStore) *API {
	return &API{Store: store}
}

func (api *API) Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/vectors", api.InsertVector)
	r.Get("/vectors/{id}", api.GetVector)
	r.Delete("/vectors/{id}", api.DeleteVector)

	return r
}

func (api *API) InsertVector(w http.ResponseWriter, r *http.Request) {
	var v store.Vector 
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	api.Store.Insert(v)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v) // Optionally return inserted Vector
}

func (api *API) GetVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := api.Store.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(v)
}

func (api *API) DeleteVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := api.Store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
