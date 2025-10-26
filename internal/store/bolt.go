package store

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.etcd.io/bbolt"
	"vectraDB/internal/models"
	"vectraDB/pkg/errors"
)

type boltStore struct {
	db     *bbolt.DB
	config Config
	mu     sync.RWMutex
	
	// In-memory cache for vectors
	vectors map[string]*models.Vector
	// Inverted index for metadata filtering
	index map[string]map[string]map[string]bool
}

func NewBoltStore(config Config) (Store, error) {
	db, err := bbolt.Open(config.DBPath, 0600, &bbolt.Options{
		Timeout: config.Timeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, http.StatusInternalServerError, "failed to open database")
	}

	store := &boltStore{
		db:      db,
		config:  config,
		vectors: make(map[string]*models.Vector),
		index:   make(map[string]map[string]map[string]bool),
	}

	// Initialize buckets
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, err
	}

	// Load vectors into memory
	if err := store.loadVectors(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *boltStore) initBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("vectors"))
		if err != nil {
			return errors.Wrap(err, http.StatusInternalServerError, "failed to create vectors bucket")
		}
		
		_, err = tx.CreateBucketIfNotExists([]byte("documents"))
		if err != nil {
			return errors.Wrap(err, http.StatusInternalServerError, "failed to create documents bucket")
		}
		
		return nil
	})
}

func (s *boltStore) loadVectors() error {
	return s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("vectors"))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var vector models.Vector
			if err := json.Unmarshal(v, &vector); err != nil {
				return errors.Wrap(err, http.StatusInternalServerError, "failed to unmarshal vector")
			}
			
			s.vectors[string(k)] = &vector
			s.addToIndex(&vector)
			return nil
		})
	})
}

func (s *boltStore) addToIndex(vector *models.Vector) {
	for key, val := range vector.Metadata {
		if _, ok := s.index[key]; !ok {
			s.index[key] = make(map[string]map[string]bool)
		}
		if _, ok := s.index[key][val]; !ok {
			s.index[key][val] = make(map[string]bool)
		}
		s.index[key][val][vector.ID] = true
	}
}

func (s *boltStore) removeFromIndex(vector *models.Vector) {
	for key, val := range vector.Metadata {
		if fieldMap, ok := s.index[key]; ok {
			if idMap, ok := fieldMap[val]; ok {
				delete(idMap, vector.ID)
				if len(idMap) == 0 {
					delete(fieldMap, val)
				}
			}
		}
	}
}

func (s *boltStore) InsertVector(ctx context.Context, vector *models.Vector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vector already exists
	if _, exists := s.vectors[vector.ID]; exists {
		return errors.ErrVectorExists
	}

	// Set timestamps
	now := time.Now()
	vector.CreatedAt = now
	vector.UpdatedAt = now

	// Marshal vector
	data, err := json.Marshal(vector)
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to marshal vector")
	}

	// Store in database
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("vectors"))
		return bucket.Put([]byte(vector.ID), data)
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to store vector")
	}

	// Update in-memory cache
	s.vectors[vector.ID] = vector
	s.addToIndex(vector)

	return nil
}

func (s *boltStore) GetVector(ctx context.Context, id string) (*models.Vector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vector, exists := s.vectors[id]
	if !exists {
		return nil, errors.ErrVectorNotFound
	}

	return vector, nil
}

func (s *boltStore) UpdateVector(ctx context.Context, id string, vector *models.Vector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vector exists
	oldVector, exists := s.vectors[id]
	if !exists {
		return errors.ErrVectorNotFound
	}

	// Remove old vector from index
	s.removeFromIndex(oldVector)

	// Set timestamps
	vector.ID = id
	vector.CreatedAt = oldVector.CreatedAt
	vector.UpdatedAt = time.Now()

	// Marshal vector
	data, err := json.Marshal(vector)
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to marshal vector")
	}

	// Update in database
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("vectors"))
		return bucket.Put([]byte(id), data)
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to update vector")
	}

	// Update in-memory cache
	s.vectors[id] = vector
	s.addToIndex(vector)

	return nil
}

func (s *boltStore) DeleteVector(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vector exists
	vector, exists := s.vectors[id]
	if !exists {
		return errors.ErrVectorNotFound
	}

	// Remove from database
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("vectors"))
		return bucket.Delete([]byte(id))
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to delete vector")
	}

	// Remove from in-memory cache
	delete(s.vectors, id)
	s.removeFromIndex(vector)

	return nil
}

func (s *boltStore) ListVectors(ctx context.Context, limit, offset int) ([]*models.Vector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vectors := make([]*models.Vector, 0, len(s.vectors))
	for _, vector := range s.vectors {
		vectors = append(vectors, vector)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(vectors) {
		return []*models.Vector{}, nil
	}
	if end > len(vectors) {
		end = len(vectors)
	}

	return vectors[start:end], nil
}

func (s *boltStore) Health(ctx context.Context) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		// Try to access the vectors bucket
		bucket := tx.Bucket([]byte("vectors"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "vectors bucket not found")
		}
		return nil
	})
}

func (s *boltStore) Close() error {
	return s.db.Close()
}
