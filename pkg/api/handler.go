package api

import (
  "encoding/json"
	"fmt"
  "net/http"

  "github.com/go-chi/chi/v5"
  "vectraDB/pkg/store"
)

type API struct {
	Store *store.VectorStore
}

type responseItem struct {
	ID 				string									`json:"id"`
	Score 			float64 								`json:"score"`
	Metadata 		map[string]interface{} 					`json:"metadata,omitempty"`
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
	r.Put("/vectors/{id}", api.UpdateVector)
	r.Delete("/vectors/{id}", api.DeleteVector)
	r.Post("/search", api.SearchVectors)

	r.Post("/hybrid/search", api.HybridSearchVector)
	return r
}

func (api *API) InsertVector(w http.ResponseWriter, r *http.Request) {
	var v store.Vector 
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if v.ID == "" {
		http.Error(w, "missing vector ID", http.StatusBadRequest)
		return
	}
	
	if err := api.Store.Insert(v); err != nil {
		http.Error(w, fmt.Sprintf("failed to insert vector: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v) // Optionally return inserted Vector
}

// Retrieves a vector by its ID
func (api *API) GetVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := api.Store.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(v)
}


func (api *API) UpdateVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var v store.Vector

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	v.ID = id
	if err := api.Store.UpdateVector(id, v); err != nil {
		http.Error(w, fmt.Sprintf("failed to update: %v", err),http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// deletes a vector by ID
func (api *API) DeleteVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := api.Store.DeleteVector(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sets resonable defaults for search queries
func (req *SearchRequest) SetDefaults() {
	if req.Weights == nil {
		req.Weights = map[string]float64{"vector": 1.0, "metadata": 0.0}
	}
	if _, ok := req.Weights["vector"]; !ok {
		req.Weights["vector"] = 1.0
	}
	if _, ok := req.Weights["metadata"]; !ok {
		req.Weights["metadata"] = 0.0
	}

	// Normalize Weights
	sum := req.Weights["vector"] + req.Weights["metadata"]
	if sum > 0 {
		req.Weights["vector"] /= sum
		req.Weights["metadata"] /= sum
	}
}

// Handles search queries
func (api *API) SearchVectors(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.SetDefaults()

	params := store.SearchParams{
	Query:   req.Query,
	TopK:    req.TopK,
	Filter:  req.Filter,
	Page:    req.Page,
	Limit:   req.Limit,
	Weights: req.Weights,
}

	results, total := api.Store.Search(params)

	resp := struct {
		Total 		int						`json:"total"`
		Page 		int						`json:"page"`
		Limit		int 					`json:"limit"`
		Results 	[]store.SearchResult	`json:"results"`
	}{
		Total: 	total,
		Page: 	req.Page,
		Limit: 	req.Limit,
		Results:results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (api *API) HybridSearchVector(w http.ResponseWriter, r *http.Request) {
	var req store.HybridSearchParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode HybridSearch request: %v", err)
		return
	}

	// Default weights
	if req.VectorWeight+req.KeywordWeight == 0 {
		req.VectorWeight = 0.5
		req.KeywordWeight = 0.5
	}

	// Default limit
	if req.Limit <= 0 {
		req.Limit = 10
	}

	results, total := api.Store.HybridSearch(req)

	resp := store.HybridSearchResponse{
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode HybridSearch response: %v", err)
	}
}
