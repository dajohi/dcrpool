package database

import (
	"encoding/binary"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
)

var (
	// PoolBkt is the main bucket of mining pool, all other buckets
	// are nested within it.
	PoolBkt = []byte("poolbkt")

	// AccountBkt stores all registered accounts for the mining pool.
	AccountBkt = []byte("accountbkt")

	// NameIdxBkt is an index of all account names mapped to their ids.
	NameIdxBkt = []byte("nameidxbkt")

	// ShareBkt stores all client shares for the mining pool.
	ShareBkt = []byte("sharebkt")

	// WorkBkt stores work submissions from the pool accepted by the network,
	// periodically pruned by the current chain tip height.
	WorkBkt = []byte("workbkt")

	// PaymentBkt stores all payments.
	PaymentBkt = []byte("paymentbkt")

	// VersionK is the key of the current version of the database.
	VersionK = []byte("version")
)

// ErrValueNotFound is returned when a provided database key does not map
// to any value.
func ErrValueNotFound(key []byte) error {
	return fmt.Errorf("associated value for key '%v' not found", string(key))
}

// ErrBucketNotFound is returned when a provided database bucket cannot be
// found.
func ErrBucketNotFound(bucket []byte) error {
	return fmt.Errorf("bucket '%v' not found", string(bucket))
}

// OpenDB creates a connection to the provided bolt storage, the returned
// connection storage should always be closed after use.
func OpenDB(storage string) (*bolt.DB, error) {
	db, err := bolt.Open(storage, 0600,
		&bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return db, nil
}

// CreateBuckets creates all storage buckets of the mining pool.
func CreateBuckets(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		pbkt := tx.Bucket(PoolBkt)
		if pbkt == nil {
			// Initial bucket layout creation.
			pbkt, err = tx.CreateBucketIfNotExists(PoolBkt)
			if err != nil {
				return fmt.Errorf("failed to create '%s' bucket: %v",
					string(PoolBkt), err)
			}

			// Persist the database version.
			vbytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(vbytes, uint32(DBVersion))
			pbkt.Put(VersionK, vbytes)
		}

		// Create all other buckets nested within.
		_, err = pbkt.CreateBucketIfNotExists(AccountBkt)
		if err != nil {
			return fmt.Errorf("failed to create '%v' bucket: %v",
				string(AccountBkt), err)
		}

		_, err = pbkt.CreateBucketIfNotExists(ShareBkt)
		if err != nil {
			return fmt.Errorf("failed to create '%v' bucket: %v",
				string(ShareBkt), err)
		}

		_, err = pbkt.CreateBucketIfNotExists(NameIdxBkt)
		if err != nil {
			return fmt.Errorf("failed to create '%v' bucket: %v",
				string(NameIdxBkt), err)
		}

		_, err = pbkt.CreateBucketIfNotExists(WorkBkt)
		if err != nil {
			return fmt.Errorf("failed to create '%v' bucket: %v",
				string(WorkBkt), err)
		}

		_, err = pbkt.CreateBucketIfNotExists(PaymentBkt)
		if err != nil {
			return fmt.Errorf("failed to create '%v' bucket: %v",
				string(PaymentBkt), err)
		}

		return nil
	})
	return err
}

// Delete removes the specified key and its associated value from the provided
// bucket.
func Delete(db *bolt.DB, bucket, key []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		pbkt := tx.Bucket(PoolBkt)
		if pbkt == nil {
			return ErrBucketNotFound(bucket)
		}
		b := pbkt.Bucket(bucket)
		return b.Delete(key)
	})
	return err
}

// GetIndexValue asserts if a an index value exists in the provided bucket.
func GetIndexValue(db *bolt.DB, bucket, key []byte) ([]byte, error) {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		pbkt := tx.Bucket(PoolBkt)
		if pbkt == nil {
			return ErrBucketNotFound(bucket)
		}
		bkt := pbkt.Bucket(bucket)
		if bkt == nil {
			return ErrBucketNotFound(bucket)
		}
		value = bkt.Get(key)
		return nil
	})
	return value, err
}

// UpdateIndex updates an index entry in the provided bucket.
func UpdateIndex(db *bolt.DB, bucket []byte, key []byte, value []byte) error {
	// Update the username index.
	err := db.Update(func(tx *bolt.Tx) error {
		pbkt := tx.Bucket(PoolBkt)
		if pbkt == nil {
			return ErrBucketNotFound(bucket)
		}
		bkt := pbkt.Bucket(bucket)
		if bkt == nil {
			return ErrBucketNotFound(bucket)
		}
		err := bkt.Put(key, value)
		return err
	})
	return err
}

// RemoveIndex deletes an index entry in the provided bucket.
func RemoveIndex(db *bolt.DB, bucket []byte, key []byte) error {
	// Update the username index.
	err := db.Update(func(tx *bolt.Tx) error {
		pbkt := tx.Bucket(PoolBkt)
		if pbkt == nil {
			return ErrBucketNotFound(bucket)
		}
		bkt := pbkt.Bucket(bucket)
		if bkt == nil {
			return ErrBucketNotFound(bucket)
		}
		err := bkt.Delete(key)
		return err
	})
	return err
}
