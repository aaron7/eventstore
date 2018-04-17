package store

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func structToBytes(s interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
