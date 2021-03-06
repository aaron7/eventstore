package db

import (
	"fmt"
	"net/url"

	badger "github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

// KeyValuePair describes a key and a value
type KeyValuePair struct {
	Key   []byte
	Value []byte
}

// DB is the interface for the database
type DB interface {
	LookupValue(key []byte) (value []byte, exists bool, err error)
	SetKeyValues([]KeyValuePair) error
	GetSequence(key []byte, bandwidth uint64) (Sequence, error)
	RangeKeys(prefix []byte, keyItr func([]byte) error) error
	Stream(prefix []byte, keyToList func(key []byte, itr *badger.Iterator) (list *pb.KVList, err error), send func(list *pb.KVList) error) error
	DropAll() error
	Close() error
}

// Sequence is the interface for uint64 sequencers
type Sequence interface {
	Next() (uint64, error)
}

// New creates a new database
func New(uri string) (DB, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	var d DB
	switch u.Scheme {
	case "badger":
		d = newBadgerDB(fmt.Sprintf("%s%s", u.Host, u.Path))
	case "memory":
		d = newMemoryDB()
	default:
		return nil, fmt.Errorf("Unknown database type: %s", u.Scheme)
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}
