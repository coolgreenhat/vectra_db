package store

import (
	"errors"
	"sync"
)

type VectorStore struct {
	vectors map[string]Vector 
	mu 			sync.RWMutex
}

func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors: make(map[string]Vector),
	}
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
