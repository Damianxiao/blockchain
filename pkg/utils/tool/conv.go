package tool

import (
	"encoding/binary"
)

func IntToBytes(n int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(n))
	return buf
}

func BytesToInt(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
