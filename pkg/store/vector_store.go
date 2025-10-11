package store

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"go.etcd.io/bbolt"
	"math"
	"errors"
)

var vectorsBucket = []byte("vectors") // bucket name

// --- Insert / Get / Delete / Update ---

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
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
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

func (s *VectorStore) DeleteVector(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vec, exists := s.vectors[id]
	if !exists {
		return fmt.Errorf("vector %s not found", id)
	}

	// Remove from index
	for key, val := range vec.Metadata {
		valStr := fmt.Sprintf("%v", val)
		if s.index[key] != nil && s.index[key][valStr] != nil {
			delete(s.index[key][valStr], id)
			if len(s.index[key][valStr]) == 0 {
				delete(s.index[key], valStr)
			}
		}
	}

	delete(s.vectors, id)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
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

// addToIndex and removeFromIndex assume metadata is map[string]string
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

func (s *VectorStore) removeFromIndex(v Vector) {
	for key, val := range v.Metadata {
		if fieldMap, ok := s.index[key]; ok {
			if idMap, ok := fieldMap[val]; ok {
				delete(idMap, v.ID)
			}
		}
	}
}

// FilterVectors unchanged from your version, safe and efficient
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

// HybridSearch using BM25Scores + fuzzyThreshold
func (s *VectorStore) HybridSearch(params HybridSearchParams) ([]HybridSearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build ordered lists: ids and docs so BM25Scores indices align
	ids := make([]string, 0, len(s.vectors))
	docs := make([]string, 0, len(s.vectors))
	for id, v := range s.vectors {
		ids = append(ids, id)
		docs = append(docs, v.Text)
	}

	// compute BM25 + fuzzy partial-credit scores
	fuzzyThreshold := params.FuzzyThreshold
	bm25Scores := BM25Scores(params.Query, docs, fuzzyThreshold)

	results := make([]HybridSearchResult, 0, len(ids))
	for i, id := range ids {
		v := s.vectors[id]

		vectorScore := 0.0
		if len(params.QueryVector) > 0 && len(v.Values) > 0 {
			if vs, err := CosineSimilarity(params.QueryVector, v.Values); err == nil {
				vectorScore = vs
			}
		}

		keywordScore := bm25Scores[i]
		hybridScore := params.VectorWeight*vectorScore + params.KeywordWeight*keywordScore

		results = append(results, HybridSearchResult{
			ID:           v.ID,
			Text:         v.Text,
			VectorScore:  vectorScore,
			KeywordScore: keywordScore,
			HybridScore:  hybridScore,
		})
	}

	// sort by hybrid score
	sort.Slice(results, func(i, j int) bool {
		return results[i].HybridScore > results[j].HybridScore
	})

	if params.Limit > 0 && len(results) > params.Limit {
		results = results[:params.Limit]
	}

	return results, nil
}
