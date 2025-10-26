package models

import (
	"time"
)

type Vector struct {
	ID       string            `json:"id" validate:"required"`
	Vector   []float64         `json:"vector" validate:"required,min=1"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type Document struct {
	ID        string    `json:"id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	Content   string    `json:"content" validate:"required"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SearchRequest struct {
	Query   []float64          `json:"query" validate:"required,min=1"`
	TopK    int                `json:"top_k" validate:"min=1,max=1000"`
	Filter  map[string]string  `json:"filter,omitempty"`
	Page    int                `json:"page,omitempty" validate:"min=1"`
	Limit   int                `json:"limit,omitempty" validate:"min=1,max=100"`
	Weights map[string]float64 `json:"weights,omitempty"`
}

type SearchResult struct {
	Vector Vector  `json:"vector"`
	Score  float64 `json:"score"`
}

type SearchResponse struct {
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Results []SearchResult `json:"results"`
}

type HybridSearchRequest struct {
	Query         string    `json:"query" validate:"required"`
	QueryVector   []float64 `json:"query_vector" validate:"required,min=1"`
	VectorWeight  float64   `json:"vector_weight" validate:"min=0,max=1"`
	KeywordWeight float64   `json:"keyword_weight" validate:"min=0,max=1"`
	FuzzyWeight   float64   `json:"fuzzy_weight" validate:"min=0,max=1"`
	Limit         int       `json:"limit" validate:"min=1,max=100"`
	Page          int       `json:"page" validate:"min=1"`
}

type HybridSearchResult struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	VectorScore  float64 `json:"vector_score"`
	KeywordScore float64 `json:"keyword_score"`
	HybridScore  float64 `json:"hybrid_score"`
}

type HybridSearchResponse struct {
	Total   int                   `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
	Results []HybridSearchResult  `json:"results"`
}

type CreateVectorRequest struct {
	ID       string            `json:"id" validate:"required"`
	Vector   []float64         `json:"vector" validate:"required,min=1"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type UpdateVectorRequest struct {
	Vector   []float64         `json:"vector" validate:"required,min=1"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CreateDocumentRequest struct {
	ID      string   `json:"id" validate:"required"`
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags,omitempty"`
}

type UpdateDocumentRequest struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags,omitempty"`
}
