package store

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/aaron7/eventstore/pkg/db"
)

var eventsCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "eventstore",
	Name:      "ingest_events_total",
	Help:      "The total number of events written.",
})

func init() {
	prometheus.MustRegister(eventsCounter)
}

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
	eventsCounter.Add(float64(len(events)))

	return s.DB.SetKeyValues(indexEntries)
}

// M ...
type M map[string]interface{}

// QueryResult contains data result
type QueryResult struct {
	Data []QueryResultData `json:"data"`
}

// QueryResultData ...
type QueryResultData struct {
	Name   string `json:"name"`
	Result []M    `json:"result"`
}

// QueryEvents takes a query and returns events
func (s *Store) QueryEvents(query Query) QueryResult {
	result := []QueryResultData{}

	for _, data := range query.Data {
		eventIDs := []uint64{}
		eventIDsMap := make(map[uint64]struct{})

		fetchedKeysMap := make(map[string]struct{})
		fetchedKeyValues := make(map[uint64]map[string]string)

		for i, filter := range data.Filters {
			currentFilterEventIDs := []uint64{}
			keys := s.DB.RangeKeys(getDimensionIndexRangeKey(filter.Key, filter.Value))
			for _, key := range keys {
				_, _, eventID := decodeDimensionIndexKey(key)

				// Save the value if we know it from the filter
				if _, ok := fetchedKeyValues[eventID]; !ok {
					fetchedKeyValues[eventID] = make(map[string]string)
				}
				fetchedKeyValues[eventID][filter.Key] = filter.Value

				if i == 0 {
					// Add first set of eventIDs directly to result
					eventIDs = append(eventIDs, eventID)
					eventIDsMap[eventID] = struct{}{}
				} else {
					// Otherwise create temporary list to use for intersect
					currentFilterEventIDs = append(currentFilterEventIDs, eventID)
				}
			}

			// If first filter, don't need to intersect
			if i == 0 {
				continue
			} else {
				eventIDs, eventIDsMap = intersect(currentFilterEventIDs, eventIDsMap)
			}

			// Record we fetched the key
			fetchedKeysMap[filter.Key] = struct{}{}
		}

		// Get the remaining key values if they were not included in the filter
		for _, dataKey := range data.Keys {
			if _, ok := fetchedKeysMap[dataKey]; !ok {
				// Not yet fetched this key, so fetch it and save the values
				keys := s.DB.RangeKeys(getPartialDimensionIndexRangeKey(dataKey))
				for _, key := range keys {
					_, value, eventID := decodeDimensionIndexKey(key)
					fetchedKeyValues[eventID][dataKey] = value
				}
				// Record we fetched the key
				fetchedKeysMap[dataKey] = struct{}{}
			}
		}

		// Create result
		dataResult := []M{}
		for _, eventID := range eventIDs {
			point := M{"id": eventID}
			for k := range fetchedKeysMap {
				if v, ok := fetchedKeyValues[eventID][k]; ok {
					point[k] = v
				}
			}
			dataResult = append(dataResult, point)
		}
		result = append(result, QueryResultData{Name: data.Name, Result: dataResult})
	}

	return QueryResult{Data: result}
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
