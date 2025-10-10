package store

import (
	"fmt"
	"sort"
	"encoding/json"
	"go.etcd.io/bbolt"
	"log"
	"errors"
)

const vectorBucket = "vectors"

func (s *VectorStore) Insert(v Vector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.vectors[v.ID] = v
	s.addToIndex(v)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		return b.Put([]byte(v.ID), data)
	})
}

func (s *VectorStore) Get(id string) (Vector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.vectors[id]
	if !ok {
		return Vector{}, errors.New("vector not found")
	}

	return v, nil
}

// Remove a vector from store and DB
func (s *VectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vec, ok := s.vectors[id]
	if !ok {
		return nil
	}

	for key, val := range vec.Metadata {
		valStr := fmt.Sprintf("%v", val)
		delete(s.index[key][valStr], id)
		if len(s.index[key][valStr]) == 0 {
			delete(s.index[key], valStr)
		}
	}

	delete(s.vectors, id)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		return b.Delete([]byte(id))
	})
}

func (s *VectorStore) UpdateVector(id string, newVec Vector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldVec, exists := s.vectors[id]
	if exists {
		for key, val := range oldVec.Metadata {
			valStr := fmt.Sprintf("%v", val)
			if s.index[key] != nil && s.index[key][valStr] != nil {
				delete(s.index[key][valStr], id)
				if len(s.index[key][valStr]) == 0 {
					delete(s.index[key], valStr)
				}
			}
		}
	}

	s.vectors[id] = newVec
	s.addToIndex(newVec)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(vectorsBucket)
		if err != nil {
			return err
		}
		data, err := json.Marshal(newVec)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}

// Add vector metadata to inverted index
func (s *VectorStore) addToIndex(v Vector) {
	for key, val := range v.Metadata {
		if _, ok := s.index[key]; !ok {
			s.index[key] = make(map[string]map[string]bool)
		}
		if _, ok := s.index[key][val]; !ok {
			s.index[key][val] = make(map[string]bool)
		}
		s.index[key][val][v.ID] = true
	}
}

// Remove vector metdata from index
func (s *VectorStore) removeFromIndex(v Vector) {
	for key, val := range v.Metadata {
		if fieldMap, ok := s.index[key]; ok {
			if idMap, ok := fieldMap[val]; ok {
				delete(idMap, v.ID)
			}
		}
	}
}

// Filter vectors based on metadata 
func (s *VectorStore) FilterVectors(filters map[string]string) []Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(filters) == 0 {
		all := make([]Vector, 0, len(s.vectors))
		for _, v := range s.vectors {
			all = append(all, v)
		}
		return all
	}

	var candidateIDs map[string]bool

	for key, val := range filters {
		valueMap, ok := s.index[key]
		if !ok {
			return nil
		}

		idSet, ok := valueMap[val]
		if !ok {
			return nil
		}

		if candidateIDs == nil {
			candidateIDs = make(map[string]bool, len(idSet))
			for id := range idSet {
				candidateIDs[id] = true
			}
		} else {
			for id := range candidateIDs {
				if !idSet[id] {
					delete(candidateIDs, id)
				}
			}
		}

		if len(candidateIDs) == 0 {
			return nil
		}
	}

	results := make([]Vector, 0, len(candidateIDs))
	for id := range candidateIDs {
		if v, ok := s.vectors[id]; ok {
			results = append(results, v)
		}
	}

	return results
}


func (s *VectorStore) Close() {
	s.db.Close()
}

// Search
func (s *VectorStore) Search(params SearchParams) ([]SearchResult, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//Filter
	filtered := s.FilterVectors(params.Filter)

	// Compute Scores
	results := make([]SearchResult, 0, len(filtered))
	for _, v := range filtered {
		score, err := CosineSimilarity(params.Query, v.Vector)
		if err != nil {
			log.Printf("Skipping %s: %v", v.ID, err)
			continue
		}
		results = append(results, SearchResult{
			Vector: v, 
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	total := len(results)

	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <=0 {
		params.Page = 1
	}

	start := (params.Page - 1) * params.Limit
	if start >= total {
		return []SearchResult{}, total
	}

	end := start + params.Limit
	if end > total {
		end = total
	}

	return results[start:end], total
}

// remove a vector by ID
func (s *VectorStore) DeleteVector(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vec, exists := s.vectors[id]
	if !exists {
		return fmt.Errorf("Vector %s not found", id)
	}

	// Remove from index
	for key, val := range vec.Metadata {
		valStr := fmt.Sprintf("%v", val)
		delete(s.index[key][valStr], id)
		if len(s.index[key][valStr]) == 0 {
			delete(s.index[key], valStr)
		}
	}

	// Remove from memory
	delete(s.vectors, id)

	// remove from DB
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("vectors"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		return b.Delete([]byte(id))
	})
}
