package store

import (
	"fmt"
	"sort"

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
			indexEntries = append(indexEntries, createEventIndexEntry(event.Tag, dimension, value, eventID))
		}
		if err != nil {
			return err
		}
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
	Name   string                 `json:"name"`
	Result []DecodedEvent         `json:"result"`
	Meta   map[string]interface{} `json:"meta"`
}

// DecodedEvent ...
type DecodedEvent struct {
	EventID uint64             `json:"eventID"`
	Tag     string             `json:"tag"`
	Data    []DecodedEventData `json:"data"`
}

// DecodedEventData ...
type DecodedEventData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// QueryEvents takes a query and returns events
func (s *Store) QueryEvents(query Query) QueryResult {
	result := []QueryResultData{}

	for _, data := range query.Data {
		// Final list of events
		var finalEvents []DecodedEvent

		// Store keys we have fetched (will be a small map)
		fetchedKeysMap := make(map[string]struct{})

		for i, filter := range data.Filters {
			events := []DecodedEvent{}
			keyItr := func(key []byte) error {
				// Benchmark: 0.33 seconds for 3.3m keys
				// TODO: Find faster decoding
				_, _, eventValue, eventID := decodeEventIndexKey(key)

				if i == 0 {
					// Benchmark: Using map is 0.6s longer. Ids is 0.3s quicker.
					// TODO: Find fasting encoding than struct?
					events = append(events, DecodedEvent{EventID: eventID, Tag: data.Tag, Data: []DecodedEventData{{filter.Key, eventValue}}})
				} else {
					// Intersect by searching the events list from previous combined filters and only
					// adding the event from this filter if it is also in the previous combined filters.
					idx := sort.Search(len(finalEvents), func(i int) bool {
						return eventID <= finalEvents[i].EventID
					})
					if idx < len(finalEvents) && finalEvents[idx].EventID == eventID {
						finalEvents[idx].Data = append(finalEvents[idx].Data, DecodedEventData{filter.Key, eventValue})
						events = append(events, finalEvents[idx])
					}
				}

				return nil
			}
			err := s.DB.RangeKeys(getPartialEventIndexValueRangeKey(data.Tag, filter.Key, filter.Value), keyItr)
			if err != nil {
				fmt.Println("error")
				break
			}

			finalEvents = events

			// Record we fetched the key
			fetchedKeysMap[filter.Key] = struct{}{}
		}

		// Get the remaining key values if they were not included in the filter
		for _, dataKey := range data.Keys {
			if _, ok := fetchedKeysMap[dataKey]; !ok {
				// Not yet fetched this key, so fetch it and save the values
				keyItr := func(key []byte) error {
					// Benchmark: 0.33 seconds for 3.3m keys
					// TODO: Find faster decoding
					_, _, eventValue, eventID := decodeEventIndexKey(key)

					// Intersect by searching the events list from previous combined filters and only
					// adding the event from this filter if it is also in the previous combined filters.
					idx := sort.Search(len(finalEvents), func(i int) bool {
						return eventID <= finalEvents[i].EventID
					})
					if idx < len(finalEvents) && finalEvents[idx].EventID == eventID {
						finalEvents[idx].Data = append(finalEvents[idx].Data, DecodedEventData{dataKey, eventValue})
					}
					return nil
				}
				s.DB.RangeKeys(getPartialEventIndexDimensionRangeKey(data.Tag, dataKey), keyItr)

				// Record we fetched the key
				fetchedKeysMap[dataKey] = struct{}{}
			}
		}

		// Apply operations
		meta := make(map[string]interface{})
		for _, operation := range data.Operations {
			if operation.Type == "count" {
				count := len(finalEvents)
				meta["count"] = count
			}
			if operation.Type == "uniqueCount" {
				var uniqueCount uint64
				uniqueMap := make(map[string]struct{})

				var keyIndex int
				if len(finalEvents) > 0 {
					for i, kv := range finalEvents[0].Data {
						if kv.Key == operation.Key {
							keyIndex = i
							break
						}
					}
				}

				for _, event := range finalEvents {
					if _, ok := uniqueMap[event.Data[keyIndex].Value]; !ok {
						uniqueCount++
						uniqueMap[event.Data[keyIndex].Value] = struct{}{}
					}
				}
				meta["uniqueCount"] = uniqueCount
			}
		}

		// Hide the event data is HideData is true
		if data.HideData {
			finalEvents = []DecodedEvent{}
		}

		result = append(result, QueryResultData{Name: data.Name, Result: finalEvents, Meta: meta})
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
