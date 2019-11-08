package db

import (
	"sync"
	"sync/atomic"

	badger "github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

// MemoryDB implements DB
type MemoryDB struct {
	db        sync.Map
	sequences map[string]*memorySequence
}

func newMemoryDB() *MemoryDB {
	return &MemoryDB{
		db:        sync.Map{},
		sequences: make(map[string]*memorySequence),
	}
}

// LookupValue implements DB
func (m *MemoryDB) LookupValue(key []byte) (value []byte, exists bool, err error) {
	val, ok := m.db.Load(string(key))
	if !ok {
		return nil, false, nil
	}
	return val.([]byte), true, nil
}

// SetKeyValues implements DB
func (m *MemoryDB) SetKeyValues(kvs []KeyValuePair) error {
	for _, kv := range kvs {
		m.db.Store(string(kv.Key), kv.Value)
	}
	return nil
}

// GetSequence implements DB
func (m *MemoryDB) GetSequence(key []byte, bandwidth uint64) (Sequence, error) {
	_, ok := m.sequences[string(key)]
	if !ok {
		initial := uint64(0)
		m.sequences[string(key)] = &memorySequence{i: &initial}
	}
	return m.sequences[string(key)], nil
}

type memorySequence struct {
	i *uint64
}

func (ms *memorySequence) Next() (uint64, error) {
	next := atomic.AddUint64(ms.i, 1)
	return next, nil
}

// RangeKeys implements DB
func (m *MemoryDB) RangeKeys(prefix []byte, keyItr func([]byte) error) error {
	return nil
}

// Close implements DB
func (m *MemoryDB) Close() error {
	return nil
}

// Stream implements DB
func (m *MemoryDB) Stream(prefix []byte, keyToList func(key []byte, itr *badger.Iterator) (list *pb.KVList, err error), send func(list *pb.KVList) error) error {
	return nil
}

// DropAll implements DB
func (m *MemoryDB) DropAll() error {
	m.db = sync.Map{}
	return nil
}
