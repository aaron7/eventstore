package store

import (
	"regexp"
	"sort"
)

// equalFilter filters the DB and merges keys equal to the value
func equalFilter(tag, key, value string, store *Store, mergeEvents []DecodedEvent, first bool) ([]DecodedEvent, error) {
	events := []DecodedEvent{}
	keyItr := func(k []byte) error {
		// Benchmark: 0.33 seconds for 3.3m keys
		// TODO: Find faster decoding
		_, _, eventValue, ts, eventID := decodeEventIndexKey(k)

		if first {
			// Benchmark: Using map is 0.6s longer. Ids is 0.3s quicker.
			// TODO: Find fasting encoding than struct?
			events = append(events, DecodedEvent{ID: eventID, TS: ts, Tag: tag, Data: []DecodedEventData{{key, eventValue}}})
		} else {
			// Intersect by searching the events list from previous combined filters and only
			// adding the event from this filter if it is also in the previous combined filters.
			idx := sort.Search(len(mergeEvents), func(i int) bool {
				return eventID <= mergeEvents[i].ID
			})
			if idx < len(mergeEvents) && mergeEvents[idx].ID == eventID {
				mergeEvents[idx].Data = append(mergeEvents[idx].Data, DecodedEventData{key, eventValue})
				events = append(events, mergeEvents[idx])
			}
		}

		return nil
	}

	err := store.DB.RangeKeys(getPartialEventIndexValueRangeKey(tag, key, value), keyItr)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// regexFilter filters the DB and merges keys equal to the value
func regexFilter(tag, key, regex string, store *Store, mergeEvents []DecodedEvent, first bool) ([]DecodedEvent, error) {
	events := []DecodedEvent{}
	keyItr := func(k []byte) error {
		// Benchmark: 0.33 seconds for 3.3m keys
		// TODO: Find faster decoding
		_, _, eventValue, ts, eventID := decodeEventIndexKey(k)

		// Do not add event if we don't match regex
		// TODO: Improve performance
		matched, _ := regexp.MatchString(regex, eventValue)
		if !matched {
			return nil
		}

		if first {
			// Benchmark: Using map is 0.6s longer. Ids is 0.3s quicker.
			// TODO: Find fasting encoding than struct?
			events = append(events, DecodedEvent{ID: eventID, TS: ts, Tag: tag, Data: []DecodedEventData{{key, eventValue}}})
		} else {
			// Intersect by searching the events list from previous combined filters and only
			// adding the event from this filter if it is also in the previous combined filters.
			idx := sort.Search(len(mergeEvents), func(i int) bool {
				return eventID <= mergeEvents[i].ID
			})
			if idx < len(mergeEvents) && mergeEvents[idx].ID == eventID {
				mergeEvents[idx].Data = append(mergeEvents[idx].Data, DecodedEventData{key, eventValue})
				events = append(events, mergeEvents[idx])
			}
		}

		return nil
	}

	err := store.DB.RangeKeys(getPartialEventIndexDimensionRangeKey(tag, key), keyItr)
	if err != nil {
		return nil, err
	}

	return events, nil
}
