package store

import (
	"fmt"
	"strings"

	"github.com/aaron7/eventstore/pkg/db"
)

// Dimension Index
// (dimension, value, event_id) => nil
// Used to lookup events based on dimension and value pair
const dimensionIndexPrefix = "d"

// TODO: Better encoding
func getDimensionIndexEntryKey(dimension, value string, eventID uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s", dimensionIndexPrefix, dimension, value, uint64ToBytes(eventID)))
}

// TODO: Better encoding
func decodeDimensionIndexKey(key []byte) (dimension, value string, eventID uint64) {
	s := string(key)
	parts := strings.SplitN(s, ":", 4)
	return parts[1], parts[2], bytesToUint64([]byte(parts[3]))
}

// TODO: Better encoding
func getDimensionIndexRangeKey(dimension, value string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:", dimensionIndexPrefix, dimension, value))
}

func createDimensionIndexEntry(dimension, value string, eventID uint64) db.KeyValuePair {
	return db.KeyValuePair{
		Key:   getDimensionIndexEntryKey(dimension, value, eventID),
		Value: nil,
	}
}

// Event Index
// (event_id) => { TS: time, SampleRate: uint64 }
// Used to lookup complete events
const eventIndexPrefix = "e"

func getEventIndexEntryKey(eventID uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s", eventIndexPrefix, uint64ToBytes(eventID)))
}

// EventMetaData is stored in the event index
type EventMetaData struct {
	TS         int
	Samplerate int
}

func createEventIndexEntry(eventID uint64, eventMetaData EventMetaData) (db.KeyValuePair, error) {
	b, err := structToBytes(eventMetaData)
	if err != nil {
		return db.KeyValuePair{}, err
	}
	return db.KeyValuePair{
		Key:   getEventIndexEntryKey(eventID),
		Value: b,
	}, nil
}
