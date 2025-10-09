package store

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

var (
	vectorsBucket = []byte("vectors")
)

func NewVectorStore(path string) (*VectorStore, error) {
	db, err := bbolt.Open(path, 0666, &bbolt.Options{Timeout: 1 * time.second})
	if err != nil {
		return nil, err
	}

	store := &VectorStore{
		vectors: make(map[string]Vector),
		index: make(map[string]map[string]map[string]bool),
		db:   db,	
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(vectorsBucket)
		return e
	})
	if err != nil {
		return nil, err
	}

	if err := store.loadFromDB(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *VectorStore) Close() error {
	return s.db.Close()
}

func (s *VectorStore) loadFromDB() error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketVectors))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var vec Vector
			if err := json.Unmarshal(v, &vec); err != nil {
				log.Printf("failed to unmarshal vector %s: %v", k, err)
				return nil
			}
			s.vectors[string(k)] = vec
			s.addToIndex(vec)
			return nil
		})
	})
}


func (s *VectorStore) Add(v Vector) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		return b.Put([]byte(v.ID), data)
	})
}

func (s *VectorStore) All() ([]Vector, error) {
	vectors := []Vector{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		return b.ForEach(func(k,v []byte) error {
			var vec Vector 
			if err := json.Unmarshal(v, &vec); err != nil {
				log.Printf("Error decoding vector %s: %v", k, err)
				return nil
			}
			vectors = append(vectors, vec)
			return nil
		})
	})
	return vectors, err
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
