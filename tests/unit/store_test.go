// Package store provides unit tests for the store package.
// All tests automatically clean up their database files after completion.
package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vectraDB/internal/models"
	"vectraDB/internal/store"
)

func cleanupTestDB(t *testing.T, dbPath string) {
	t.Cleanup(func() {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to cleanup test database %s: %v", dbPath, err)
		}
	})
}

func cleanupAllTestDBs(t *testing.T) {
	t.Cleanup(func() {
		// Clean up any remaining test database files
		pattern := "test_*.db"
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Logf("Failed to find test database files: %v", err)
			return
		}
		
		for _, match := range matches {
			if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
				t.Logf("Failed to cleanup test database %s: %v", match, err)
			}
		}
	})
}

func TestBoltStore_InsertVector(t *testing.T) {
	cleanupAllTestDBs(t)
	dbPath := "test_insert_" + t.Name() + ".db"
	cleanupTestDB(t, dbPath)
	
	// Create a temporary store for testing
	testStore, err := store.NewBoltStore(store.Config{
		DBPath:   dbPath,
		Timeout:  1 * time.Second,
		MaxConns: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer testStore.Close()

	// Test vector
	vector := &models.Vector{
		ID:     "test-vector-1",
		Vector: []float64{0.1, 0.2, 0.3, 0.4},
		Text:   "Test vector",
		Metadata: map[string]string{
			"category": "test",
			"source":   "unit-test",
		},
	}

	// Insert vector
	err = testStore.InsertVector(context.Background(), vector)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Retrieve vector
	retrieved, err := testStore.GetVector(context.Background(), vector.ID)
	if err != nil {
		t.Fatalf("Failed to get vector: %v", err)
	}

	// Verify vector data
	if retrieved.ID != vector.ID {
		t.Errorf("Expected ID %s, got %s", vector.ID, retrieved.ID)
	}

	if len(retrieved.Vector) != len(vector.Vector) {
		t.Errorf("Expected vector length %d, got %d", len(vector.Vector), len(retrieved.Vector))
	}

	for i, val := range retrieved.Vector {
		if val != vector.Vector[i] {
			t.Errorf("Expected vector[%d] %f, got %f", i, vector.Vector[i], val)
		}
	}

	if retrieved.Text != vector.Text {
		t.Errorf("Expected text %s, got %s", vector.Text, retrieved.Text)
	}

	if len(retrieved.Metadata) != len(vector.Metadata) {
		t.Errorf("Expected metadata length %d, got %d", len(vector.Metadata), len(retrieved.Metadata))
	}

	for key, val := range retrieved.Metadata {
		if val != vector.Metadata[key] {
			t.Errorf("Expected metadata[%s] %s, got %s", key, vector.Metadata[key], val)
		}
	}
}

func TestBoltStore_UpdateVector(t *testing.T) {
	cleanupAllTestDBs(t)
	dbPath := "test_update_" + t.Name() + ".db"
	cleanupTestDB(t, dbPath)
	
	// Create a temporary store for testing
	testStore, err := store.NewBoltStore(store.Config{
		DBPath:   dbPath,
		Timeout:  1 * time.Second,
		MaxConns: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer testStore.Close()

	// Insert initial vector
	vector := &models.Vector{
		ID:     "test-vector-update",
		Vector: []float64{0.1, 0.2, 0.3, 0.4},
		Text:   "Original text",
		Metadata: map[string]string{
			"category": "test",
		},
	}

	err = testStore.InsertVector(context.Background(), vector)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Update vector
	updatedVector := &models.Vector{
		ID:     "test-vector-update",
		Vector: []float64{0.5, 0.6, 0.7, 0.8},
		Text:   "Updated text",
		Metadata: map[string]string{
			"category": "updated",
			"source":   "unit-test",
		},
	}

	err = testStore.UpdateVector(context.Background(), vector.ID, updatedVector)
	if err != nil {
		t.Fatalf("Failed to update vector: %v", err)
	}

	// Retrieve updated vector
	retrieved, err := testStore.GetVector(context.Background(), vector.ID)
	if err != nil {
		t.Fatalf("Failed to get vector: %v", err)
	}

	// Verify updated data
	if retrieved.Text != updatedVector.Text {
		t.Errorf("Expected text %s, got %s", updatedVector.Text, retrieved.Text)
	}

	if len(retrieved.Metadata) != len(updatedVector.Metadata) {
		t.Errorf("Expected metadata length %d, got %d", len(updatedVector.Metadata), len(retrieved.Metadata))
	}
}

func TestBoltStore_DeleteVector(t *testing.T) {
	cleanupAllTestDBs(t)
	dbPath := "test_delete_" + t.Name() + ".db"
	cleanupTestDB(t, dbPath)
	
	// Create a temporary store for testing
	testStore, err := store.NewBoltStore(store.Config{
		DBPath:   dbPath,
		Timeout:  1 * time.Second,
		MaxConns: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer testStore.Close()

	// Insert vector
	vector := &models.Vector{
		ID:     "test-vector-delete",
		Vector: []float64{0.1, 0.2, 0.3, 0.4},
		Text:   "Test vector",
	}

	err = testStore.InsertVector(context.Background(), vector)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Delete vector
	err = testStore.DeleteVector(context.Background(), vector.ID)
	if err != nil {
		t.Fatalf("Failed to delete vector: %v", err)
	}

	// Try to retrieve deleted vector
	_, err = testStore.GetVector(context.Background(), vector.ID)
	if err == nil {
		t.Error("Expected error when retrieving deleted vector")
	}
}

func TestBoltStore_Health(t *testing.T) {
	cleanupAllTestDBs(t)
	dbPath := "test_health_" + t.Name() + ".db"
	cleanupTestDB(t, dbPath)
	
	// Create a temporary store for testing
	testStore, err := store.NewBoltStore(store.Config{
		DBPath:   dbPath,
		Timeout:  1 * time.Second,
		MaxConns: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer testStore.Close()

	// Test health check
	err = testStore.Health(context.Background())
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}
