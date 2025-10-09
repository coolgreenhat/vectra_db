package store

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"log"
)

const vectorBucket = "vectors"

func (s *VectorStore) Upsert(v Vector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old vector from Index
	if old, ok := s.vectors[v.ID]; ok {
		s.removeFromIndex(old)
	}
	
	// Save to BoltDB
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketVectors))
		data, _ := json.Marshal(v)
		return b.Put([]byte(v.ID), data)
	})

	if err != nil {
		return err
	}

	s.vectors[v.ID] = v
	s.addToIndex(v)

	return nil

}

// Remove a vector from store and DB
func (s *VectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vectors[id]
	if !ok {
		return nil
	}

	s.removeFromIndex(v)
	delete(s.vectors, id)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketVectors))
		return b.Delete([]byte(id))
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
			candidateIDs = make(map[string]bool)
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
	}

	results := []Vectors{}
	for id: = range candidateIDs {
		results = append(results, s.vectors[id])
	}
	return results
}


func (s *VectorStore) Close() {
	s.db.Close()
}

func (s *VectorStore) Search(params SearchParams) ([]SearchResult, int) {
	filtered := s.FilterVectors(params.Filter)
	results := []SearchResult{}

	for _, v := range filtered {
		score, err := CosineSimilarity(params.Query, v.Vector)
		if err != nil {
			log.Printf("Skipping Vector %s due to error : %v", v.ID, err)
			continue
		}
		results = append(results, SearchResult{Vector: v, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].score
	}

	total := len(results)
	start := (params.Page - 1) * params.Limit
	end := start + params.Limit

	if start > total {
		start = total
	}

	if end > total {
			end = total
	}

	return results[start:end], total
}
