package store

import (
	"errors"
	// "sync"
	"sort"
	"log"
)


func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors: make(map[string]Vector),
	}
}

func (s *VectorStore) All() []Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vectors := make([]Vector, 0, len(s.vectors))
	for _, v := range s.vectors {
		vectors = append(vectors, v)
	}
	return vectors
}

func (s *VectorStore) Insert(v Vector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vectors[v.ID] = v
}

func (s *VectorStore) Get(id string)(Vector, error){
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.vectors[id]
	if !ok {
		return Vector{}, errors.New("vector not found")
	}
	return v, nil
}

func (s *VectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vectors[id]; !ok {
		return errors.New("vector not found")
	}
	delete(s.vectors, id)
	return nil
}

func (s *VectorStore) List() []Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Vector, 0, len(s.vectors))
	for _, v := range s.vectors {
		list = append(list, v)
	}
	return list
}

func (s *VectorStore) Search(params SearchParams) ([]SearchResult, int) {
	vectors := s.All() // fetch all vectors

	// s.mu.RLock()
	// defer s.mu.RUnlock()

	// 1. Filter
	filtered := []Vector{}
	for _, v := range vectors {
		match := true
		for key, val := range params.Filter {
			if v.Metadata[key] != val {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, v)
		}
	}
	
	// 2. Compute Scores
	results := []SearchResult{}
	for _, v := range filtered {
		score, error := CosineSimilarity(params.Query, v.Values)
		if error != nil {
			log.Printf("Skipping vector %s due to error : %v", v.ID, error)
			continue
		}

		results = append(results, SearchResult{
			Vector: v,
			Score: score,
		})
	}

	// 3. Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 4. Pagination
	start := (params.Page - 1) * params.Limit
	if start < 0 {
		start = 0
	}
	end := start + params.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], len(results)
}
