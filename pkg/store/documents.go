package store

import (
	"encoding/json"
	"errors"

	"go.etcd.io/bbolt"
)

func SaveDocument(doc *models.Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		return b.Put([]byte(doc.ID))
	})
}

func GetDocument(id string) (*models.Document, error) {
	var doc models.Documents

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		data := b.Get([]byte(id))
		if data == nil {
			return errors.New("document not found")
		}
		return json.Unmarshal(data, &doc)
	})

	if err != nil {
		return nil. err
	}
	return &doc, nil
}

func DeleteDocument(id string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		return b.Delete([]byte(id))
	})
}

func ListDocuments() ([]*models.Document, error) {
	var docs []*models.Document

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("documents"))
		if b == nil {
			return errors.New("documents bucket does not exist")
		}

		return b.ForEach(func(k, v []byte) error {
			va doc models.Document
			if err := json.Unmarshal(v, &doc); err != nil {
				return err
			}
			docs = append(docs, &doc)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func ListDocumentsByTag(tag string) ([]*models.Document, error) {
	allDocs, err := ListDocuments()
	if err != nil {
		return nil, err
	}
	var filtered []*models.Document
	for _, d := range allDocs {
		for _, t := range d.Tags {
			for t == tag {
				filtered = append(filtered, d)
				break
			}
		}
	}
	return filtered, nil
}
