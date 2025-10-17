package store

import (
	"sync"

	"go.etcd.io/bbolt"
)

type Vector struct {
	ID 		string				`json:"id"`
	Vector		[]float64			`json:"vector"`
	Text 		string				`json:"text"`
	Metadata 	map[string]string		`json:"metadata"`
}

// Query/Search Request params
type SearchParams struct {
	Query   	[]float64				`json:"query"`
	TopK 		int					`json:"top_k"`
	Filter		map[string]string			`json:"filter,omitempty"`
	Page		int					`json:"page, omitempty"`
	Limit		int					`json:"limit, omitempty"`
	Weights 	map[string]float64			`json:"weights"`
}

type SearchResult struct {
	Vector 	Vector		`json:"vector"` 
	Score 	float64		`json:"score"`
}

type VectorStore struct {
	vectors map[string]Vector				// in-memory cache
	index 	map[string]map[string]map[string]bool // inverted index for filters
	db 			*bbolt.DB		// persistant BoltDB storage
	mu 			sync.RWMutex
}

type HybridSearchParams struct {
	Query 		string 		`json:"query"`
	QueryVector	[]float64	`json:"query_vector"`
	VectorWeight	float64		`json:"vector_weight"`
	KeywordWeight	float64		`json:"keyword_weight"`
	FuzzyWeight	float64
	Limit		int			`json:"limit"`
}

type HybridSearchResult struct {
	ID 				string 		`json:"id"`
	Text			string		`json:"text"`
	VectorScore		float64		`json:"vector_score"`
	KeywordScore	float64		`json:"keyword_score"`
	HybridScore		float64		`json:"hybrid_score"`
}

type Document struct {
	ID	string	`json:"id"`
	Title	string 	`json:"title"`
	Content	string	`json:"content"`
	Tags	[]string	`json:"tags, omitempty"`
	CreatedAt time.Time	`json:"created_at"`
}
