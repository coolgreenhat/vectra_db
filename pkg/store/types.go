package store

import (
	"sync"
)

type VectorStore struct {
	vectors map[string]Vector 
	mu 			sync.RWMutex
}

type SearchParams struct {
	Query   []float64
	TopK 		int
	Filter	map[string]string
	Page		int
	Limit		int
}

type SearchResult struct {
	Vector Vector 
	Score float64
}
