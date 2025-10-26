package store

import (
	"context"
	"time"

	"vectraDB/internal/models"
)

type VectorStore interface {
	// Vector operations
	InsertVector(ctx context.Context, vector *models.Vector) error
	GetVector(ctx context.Context, id string) (*models.Vector, error)
	UpdateVector(ctx context.Context, id string, vector *models.Vector) error
	DeleteVector(ctx context.Context, id string) error
	ListVectors(ctx context.Context, limit, offset int) ([]*models.Vector, error)
	
	// Search operations
	SearchVectors(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	HybridSearch(ctx context.Context, req *models.HybridSearchRequest) (*models.HybridSearchResponse, error)
	
	// Health check
	Health(ctx context.Context) error
	
	// Close the store
	Close() error
}

type DocumentStore interface {
	// Document operations
	InsertDocument(ctx context.Context, doc *models.Document) error
	GetDocument(ctx context.Context, id string) (*models.Document, error)
	UpdateDocument(ctx context.Context, id string, doc *models.Document) error
	DeleteDocument(ctx context.Context, id string) error
	ListDocuments(ctx context.Context, limit, offset int) ([]*models.Document, error)
	ListDocumentsByTag(ctx context.Context, tag string, limit, offset int) ([]*models.Document, error)
	
	// Health check
	Health(ctx context.Context) error
	
	// Close the store
	Close() error
}

type Store interface {
	VectorStore
	DocumentStore
}

type Config struct {
	DBPath    string
	Timeout   time.Duration
	MaxConns  int
	BatchSize int
}
