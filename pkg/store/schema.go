package store

import (
	"fmt"
	"strings"

	"github.com/aaron7/eventstore/pkg/db"
)

// Event index
// (tag, dimension, value, event_id) => nil
const eventIndexPrefix = "e"

func createEventIndexEntry(tag, dimension, value string, ts, eventID uint64) db.KeyValuePair {
	return db.KeyValuePair{
		Key:   getEventIndexEntryKey(tag, dimension, value, ts, eventID),
		Value: nil,
	}
}

func getEventIndexEntryKey(tag, dimension, value string, ts, eventID uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", eventIndexPrefix, tag, dimension, value, uint64ToBytes(ts), uint64ToBytes(eventID)))
}

func decodeEventIndexKey(key []byte) (tag, dimension, value string, ts, eventID uint64) {
	s := string(key)
	parts := strings.SplitN(s, ":", 6)
	return parts[1], parts[2], parts[3], bytesToUint64([]byte(parts[4])), bytesToUint64([]byte(parts[5]))
}

func getPartialEventIndexTagRangeKey(tag string) []byte {
	return []byte(fmt.Sprintf("%s:%s:", eventIndexPrefix, tag))
}

func getPartialEventIndexDimensionRangeKey(tag, dimension string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:", eventIndexPrefix, tag, dimension))
}

func getPartialEventIndexValueRangeKey(tag, dimension, value string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s:", eventIndexPrefix, tag, dimension, value))
}

// TODO: Use the below
// func encodeKey(ss ...[]byte) []byte {
// 	length := 0
// 	for _, s := range ss {
// 		length += len(s) + 1
// 	}
// 	output, i := make([]byte, length, length), 0
// 	for _, s := range ss {
// 		copy(output[i:i+len(s)], s)
// 		i += len(s) + 1
// 	}
// 	return output
// }

// func decodeKey(value []byte) [][]byte {
// 	components := make([][]byte, 0, 5)
// 	i, j := 0, 0
// 	for j < len(value) {
// 		if value[j] != 0 {
// 			j++
// 			continue
// 		}
// 		components = append(components, value[i:j])
// 		j++
// 		i = j
// 	}
// 	return components
// }
