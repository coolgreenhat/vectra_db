package store

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"vectraDB/internal/models"
	"vectraDB/pkg/errors"
)

func (s *boltStore) SearchVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate request
	if len(req.Query) == 0 {
		return nil, errors.ErrEmptyQuery
	}

	// Set defaults
	if req.TopK <= 0 {
		req.TopK = 10
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// Filter vectors based on metadata
	candidates := s.filterVectors(req.Filter)
	if len(candidates) == 0 {
		return &models.SearchResponse{
			Total:   0,
			Page:    req.Page,
			Limit:   req.Limit,
			Results: []models.SearchResult{},
		}, nil
	}

	// Calculate similarity scores
	results := make([]models.SearchResult, 0, len(candidates))
	for _, vector := range candidates {
		score, err := cosineSimilarity(req.Query, vector.Vector)
		if err != nil {
			continue // Skip invalid vectors
		}

		results = append(results, models.SearchResult{
			Vector: *vector,
			Score:  score,
		})
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply top-k limit
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}

	// Apply pagination
	total := len(results)
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if start >= total {
		results = []models.SearchResult{}
	} else {
		if end > total {
			end = total
		}
		results = results[start:end]
	}

	return &models.SearchResponse{
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Results: results,
	}, nil
}

func (s *boltStore) HybridSearch(ctx context.Context, req *models.HybridSearchRequest) (*models.HybridSearchResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate request
	if req.Query == "" {
		return nil, errors.ErrEmptyQuery
	}
	if len(req.QueryVector) == 0 {
		return nil, errors.ErrEmptyQuery
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.VectorWeight+req.KeywordWeight == 0 {
		req.VectorWeight = 0.5
		req.KeywordWeight = 0.5
	}

	// Get all vectors
	vectors := make([]*models.Vector, 0, len(s.vectors))
	for _, vector := range s.vectors {
		vectors = append(vectors, vector)
	}

	if len(vectors) == 0 {
		return &models.HybridSearchResponse{
			Total:   0,
			Page:    req.Page,
			Limit:   req.Limit,
			Results: []models.HybridSearchResult{},
		}, nil
	}

	// Calculate BM25 scores for keyword search
	texts := make([]string, len(vectors))
	for i, vector := range vectors {
		texts[i] = vector.Text
	}
	bm25Scores := s.calculateBM25Scores(req.Query, texts)

	// Calculate hybrid scores
	results := make([]models.HybridSearchResult, 0, len(vectors))
	for i, vector := range vectors {
		// Calculate vector similarity
		vectorScore := 0.0
		if len(vector.Vector) > 0 {
			if score, err := cosineSimilarity(req.QueryVector, vector.Vector); err == nil {
				vectorScore = score
			}
		}

		// Get keyword score
		keywordScore := bm25Scores[i]

		// Calculate hybrid score
		hybridScore := req.VectorWeight*vectorScore + req.KeywordWeight*keywordScore

		results = append(results, models.HybridSearchResult{
			ID:           vector.ID,
			Text:         vector.Text,
			VectorScore:  vectorScore,
			KeywordScore: keywordScore,
			HybridScore:  hybridScore,
		})
	}

	// Sort by hybrid score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].HybridScore > results[j].HybridScore
	})

	// Apply pagination
	total := len(results)
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if start >= total {
		results = []models.HybridSearchResult{}
	} else {
		if end > total {
			end = total
		}
		results = results[start:end]
	}

	return &models.HybridSearchResponse{
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Results: results,
	}, nil
}

func (s *boltStore) filterVectors(filters map[string]string) []*models.Vector {
	if len(filters) == 0 {
		// Return all vectors
		vectors := make([]*models.Vector, 0, len(s.vectors))
		for _, vector := range s.vectors {
			vectors = append(vectors, vector)
		}
		return vectors
	}

	// Find candidate IDs using inverted index
	var candidateIDs map[string]bool
	for key, val := range filters {
		valueMap, ok := s.index[key]
		if !ok {
			return []*models.Vector{} // No vectors match this filter
		}
		idSet, ok := valueMap[val]
		if !ok {
			return []*models.Vector{} // No vectors match this filter
		}

		if candidateIDs == nil {
			candidateIDs = make(map[string]bool, len(idSet))
			for id := range idSet {
				candidateIDs[id] = true
			}
		} else {
			// Intersect with existing candidates
			for id := range candidateIDs {
				if !idSet[id] {
					delete(candidateIDs, id)
				}
			}
		}

		if len(candidateIDs) == 0 {
			return []*models.Vector{} // No vectors match all filters
		}
	}

	// Convert candidate IDs to vectors
	vectors := make([]*models.Vector, 0, len(candidateIDs))
	for id := range candidateIDs {
		if vector, ok := s.vectors[id]; ok {
			vectors = append(vectors, vector)
		}
	}

	return vectors
}

func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same length")
	}

	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0, fmt.Errorf("zero-length vector")
	}

	return dot / (math.Sqrt(magA) * math.Sqrt(magB)), nil
}

func (s *boltStore) calculateBM25Scores(query string, texts []string) []float64 {
	queryTerms := s.tokenize(query)
	if len(queryTerms) == 0 {
		return make([]float64, len(texts))
	}

	// Calculate document frequencies
	docFreqs := make([]map[string]int, len(texts))
	termDocCount := make(map[string]int)
	totalLen := 0

	for i, text := range texts {
		tokens := s.tokenize(text)
		totalLen += len(tokens)

		freq := make(map[string]int)
		seen := make(map[string]bool)
		for _, token := range tokens {
			freq[token]++
			if !seen[token] {
				termDocCount[token]++
				seen[token] = true
			}
		}
		docFreqs[i] = freq
	}

	// Calculate average document length
	avgDocLen := float64(totalLen) / float64(len(texts))
	if len(texts) == 0 {
		avgDocLen = 0
	}

	// Calculate BM25 scores
	scores := make([]float64, len(texts))
	N := float64(len(texts))

	for i, text := range texts {
		freq := docFreqs[i]
		tokens := s.tokenize(text)
		docLen := float64(len(tokens))
		score := 0.0

		for _, term := range queryTerms {
			tf := float64(freq[term])
			if tf == 0 {
				continue
			}

			df := float64(termDocCount[term])
			if df == 0 {
				continue
			}

			// BM25 formula
			idf := math.Log(1.0 + (N-df+0.5)/(df+0.5))
			norm := tf * (1.5 + 1.0) / (tf + 1.5*(1.0-0.75+0.75*(docLen/avgDocLen)))
			score += idf * norm
		}

		scores[i] = score
	}

	return scores
}

func (s *boltStore) tokenize(text string) []string {
	parts := strings.Fields(strings.ToLower(text))
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ".,!?\"'()[]{}:;")
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}
