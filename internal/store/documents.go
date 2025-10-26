package store

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.etcd.io/bbolt"
	"vectraDB/internal/models"
	"vectraDB/pkg/errors"
)

func (s *boltStore) InsertDocument(ctx context.Context, doc *models.Document) error {
	// Check if document already exists
	existing, err := s.GetDocument(ctx, doc.ID)
	if err == nil && existing != nil {
		return errors.ErrDocumentExists
	}

	// Set timestamps
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	// Marshal document
	data, err := json.Marshal(doc)
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to marshal document")
	}

	// Store in database
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}
		return bucket.Put([]byte(doc.ID), data)
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to store document")
	}

	return nil
}

func (s *boltStore) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	var doc models.Document

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}

		data := bucket.Get([]byte(id))
		if data == nil {
			return errors.ErrDocumentNotFound
		}

		return json.Unmarshal(data, &doc)
	})
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (s *boltStore) UpdateDocument(ctx context.Context, id string, doc *models.Document) error {
	// Check if document exists
	existing, err := s.GetDocument(ctx, id)
	if err != nil {
		return err
	}

	// Set timestamps
	doc.ID = id
	doc.CreatedAt = existing.CreatedAt
	doc.UpdatedAt = time.Now()

	// Marshal document
	data, err := json.Marshal(doc)
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to marshal document")
	}

	// Update in database
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}
		return bucket.Put([]byte(id), data)
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to update document")
	}

	return nil
}

func (s *boltStore) DeleteDocument(ctx context.Context, id string) error {
	// Check if document exists
	_, err := s.GetDocument(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}
		return bucket.Delete([]byte(id))
	})
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError, "failed to delete document")
	}

	return nil
}

func (s *boltStore) ListDocuments(ctx context.Context, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}

		cursor := bucket.Cursor()
		count := 0
		skipped := 0

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			// Skip until we reach the offset
			if skipped < offset {
				skipped++
				continue
			}

			// Stop if we've reached the limit
			if count >= limit {
				break
			}

			var doc models.Document
			if err := json.Unmarshal(v, &doc); err != nil {
				continue // Skip invalid documents
			}

			documents = append(documents, &doc)
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (s *boltStore) ListDocumentsByTag(ctx context.Context, tag string, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return errors.New(http.StatusInternalServerError, "documents bucket not found")
		}

		cursor := bucket.Cursor()
		count := 0
		skipped := 0

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var doc models.Document
			if err := json.Unmarshal(v, &doc); err != nil {
				continue // Skip invalid documents
			}

			// Check if document has the specified tag
			hasTag := false
			for _, docTag := range doc.Tags {
				if docTag == tag {
					hasTag = true
					break
				}
			}

			if !hasTag {
				continue
			}

			// Skip until we reach the offset
			if skipped < offset {
				skipped++
				continue
			}

			// Stop if we've reached the limit
			if count >= limit {
				break
			}

			documents = append(documents, &doc)
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return documents, nil
}
