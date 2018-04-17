package store

import (
	"github.com/aaron7/eventstore/pkg/db"
)

// Store stores events
type Store struct {
	DB              db.DB
	EventIDSequence db.Sequence
}

// New creates a new store
func New(db db.DB) (*Store, error) {
	eventIDSequence, err := db.GetSequence([]byte("test"), 1000)
	if err != nil {
		return nil, err
	}

	return &Store{
		DB:              db,
		EventIDSequence: eventIDSequence,
	}, nil
}

// IngestEvents takes events and stores them
func (s *Store) IngestEvents(events []Event) error {
	var indexEntries []db.KeyValuePair

	for _, event := range events {
		eventID, err := s.EventIDSequence.Next()
		if err != nil {
			return err
		}
		for dimension, value := range event.Data {
			indexEntries = append(indexEntries, createDimensionIndexEntry(dimension, value, eventID))
		}
		e, err := createEventIndexEntry(eventID, EventMetaData{
			TS:         event.TS,
			Samplerate: event.Samplerate,
		})
		if err != nil {
			return err
		}
		indexEntries = append(indexEntries, e)
	}

	return s.DB.SetKeyValues(indexEntries)
}

// QueryEvents takes a query and returns events
func (s *Store) QueryEvents(query Query) []uint64 {
	eventIDs := []uint64{}
	eventIDsMap := make(map[uint64]struct{})

	for i, filter := range query.Filters {
		currentFilterEventIDs := []uint64{}
		keys := s.DB.RangeKeys(getDimensionIndexRangeKey(filter.Dimension, filter.Value))
		for _, key := range keys {
			_, _, eventID := decodeDimensionIndexKey(key)

			if i == 0 {
				// Add first set of eventIDs directly to result
				eventIDs = append(eventIDs, eventID)
				eventIDsMap[eventID] = struct{}{}
			} else {
				// Otherwise create temporary list to use for intersect
				currentFilterEventIDs = append(currentFilterEventIDs, eventID)
			}
		}

		if i == 0 {
			continue
		} else {
			eventIDs, eventIDsMap = intersect(currentFilterEventIDs, eventIDsMap)
		}
	}

	return eventIDs
}

func intersect(smallerList []uint64, largerListMap map[uint64]struct{}) ([]uint64, map[uint64]struct{}) {
	result := []uint64{}
	resultMap := make(map[uint64]struct{})
	for _, a := range smallerList {
		if _, ok := largerListMap[a]; ok {
			result = append(result, a)
			resultMap[a] = struct{}{}
		}
	}
	return result, resultMap
}
