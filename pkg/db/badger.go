package db

import (
	"log"

	"github.com/dgraph-io/badger"
)

// BadgerDB implements DB
type BadgerDB struct {
	db *badger.DB
}

func newBadgerDB(dir string) *BadgerDB {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerDB{
		db: db,
	}
}

// LookupValue implements DB
func (b *BadgerDB) LookupValue(key []byte) (value []byte, exists bool, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		copy(value, val)
		return nil
	})
	if err == badger.ErrKeyNotFound {
		return nil, false, nil
	}
	return
}

// SetKeyValues implements DB
func (b *BadgerDB) SetKeyValues(kvs []KeyValuePair) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, kv := range kvs {
			err := txn.Set(kv.Key, kv.Value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetSequence implements DB
func (b *BadgerDB) GetSequence(key []byte, bandwidth uint64) (Sequence, error) {
	return b.db.GetSequence(key, bandwidth)
}

// RangeKeys implements DB
func (b *BadgerDB) RangeKeys(prefix []byte) [][]byte {
	keys := [][]byte{}
	b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			dst := make([]byte, len(key), (cap(key)+1)*2)
			copy(dst, key)
			keys = append(keys, dst)
		}
		return nil
	})
	return keys
}

// Close implements DB
func (b *BadgerDB) Close() error {
	return b.db.Close()
}
