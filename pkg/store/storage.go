package store

import (
	"log"
	"go.etcd.io/bbolt"
)

va db *bbolt.DB

func InitDB(path string) error {
	var err error
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("vectors"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBycketIfNotExists([]byte("documents"))
		if err != nil {
			return err
		}

		return nil
	})
}
