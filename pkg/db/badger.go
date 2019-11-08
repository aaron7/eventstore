package db

import (
	"context"
	"log"

	badger "github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

// BadgerDB implements DB
type BadgerDB struct {
	db *badger.DB
}

func newBadgerDB(dir string) *BadgerDB {
	opts := badger.DefaultOptions(dir)
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
		err = item.Value(func(val []byte) error {
			copy(value, val)
			return nil
		})
		if err != nil {
			return err
		}
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
func (b *BadgerDB) RangeKeys(prefix []byte, keyItr func([]byte) error) error {
	b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keyItr(item.Key())
		}
		return nil
	})
	return nil
}

// Stream implements DB
func (b *BadgerDB) Stream(prefix []byte, keyToList func(key []byte, itr *badger.Iterator) (list *pb.KVList, err error), send func(list *pb.KVList) error) error {
	stream := b.db.NewStream()
	stream.NumGo = 16                     // Set number of goroutines to use for iteration.
	stream.Prefix = prefix                // Leave nil for iteration over the whole DB.
	stream.LogPrefix = "Badger.Streaming" // For identifying stream logs. Outputs to Logger.
	stream.KeyToList = keyToList
	stream.Send = send

	if err := stream.Orchestrate(context.Background()); err != nil {
		return err
	}

	return nil
}

// Close implements DB
func (b *BadgerDB) Close() error {
	return b.db.Close()
}
