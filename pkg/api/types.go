package api

type SearchRequest struct {
	Query   []float64          `json:"query"`
	TopK    int                `json:"top_k"`
	Filter  map[string]string  `json:"filter,omitempty"`
	Page    int                `json:"page,omitempty"`
	Limit   int                `json:"limit,omitempty"`
	Weights map[string]float64 `json:"weights,omitempty"`
}


