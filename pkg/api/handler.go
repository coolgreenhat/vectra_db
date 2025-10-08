package api

import (
  "encoding/json"
	// "math"
  "net/http"

  "github.com/go-chi/chi/v5"
  "vectraDB/pkg/store"
)

type API struct {
	Store *store.VectorStore
}

type responseItem struct {
	ID 				string									`json:"id"`
	Score 		float64 								`json:"score"`
	Metadata 	map[string]interface{}	`json:"metadata, omitempty"`
}

type searchResponse struct {
	Total 			int 					 `json:"total"` 
	Page 				int						 `json:"page"`
	Limit 			int 					 `json:"limit"`
	Results 		[]responseItem `json:"results"`
}



func NewAPI(store *store.VectorStore) *API {
	return &API{Store: store}
}

func (api *API) Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/vectors", api.InsertVector)
	r.Get("/vectors/{id}", api.GetVector)
	r.Delete("/vectors/{id}", api.DeleteVector)
	r.Post("/search", api.SearchVectors)
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

func (req *SearchRequest) SetDefaults() {
	if req.TopK <=0 {
		req.TopK = 5
	}

	if req.Limit == 0 {
		req.Limit = req.TopK
	}

	if req.Page == 0 {
		req.Page = 1
	}
}

func (api *API) SearchVectors(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.SetDefaults()

	params := store.SearchParams{
		Query: 	req.Query,
		TopK:	 	req.TopK,
		Filter: req.Filter,
		Page: 	req.Page,
		Limit: 	req.Limit,
	}

	results, total := api.Store.Search(params)

	resp := searchResponse{
		Total: 		total,
		Page: 		req.Page,
		Results: 	[]responseItem{},
	}

	for _, r := range results {
		meta := make(map[string]interface{})
		for k, v := range r.Vector.Metadata {
			meta[k] = v
		}
		resp.Results = append(resp.Results, responseItem{
			ID:			r.Vector.ID,
			Score: 	r.Score,
			Metadata: meta,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
