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
	db, err := bbolt.Open(path, 0666, &bbolt.Options{Timeout: 1 * time.Second})
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

	// Load all vectors from DB into memory
	if err := store.loadFromDB(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *VectorStore) loadFromDB() error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(vectorsBucket))
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

	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[v.ID] = v
	s.addToIndex(v)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		return b.Put([]byte(v.ID), data)
	})
}

func (s *VectorStore) All() ([]Vector, error) {
	vectors := []Vector{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(vectorsBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v[]byte) error {
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

// func (s *VectorStore) Get(id string) (Vector, error) {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
//
// 	v, ok := s.vectors[id]
// 	if !ok {
// 		return Vector{}, errors.New("vector not found")
// 	}
// 	return v, nil
// }

func (s *VectorStore) List() []Vector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]Vector, 0, len(s.vectors))
	for _, v := range s.vectors {
		list = append(list, v)
	}
	return list
}
