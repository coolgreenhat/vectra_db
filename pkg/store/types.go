package store

import (
	"sync"

	"go.etcd.io/bbolt"
)

type Vector struct {
	ID 				string							`json:"id"`
	Vector		[]float64						`json:"vector"`
	Metadata 	map[string]string		`json:"metadata"`
}

// Query/Search Request params
type SearchParams struct {
	Query   []float64						`json:"query"`
	TopK 		int									`json:"top_k"`
	Filter	map[string]string		`json:"filter,omitempty"`
	Page		int									`json:"page, omitempty"`
	Limit		int									`json:"limit, omitempty"`
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


