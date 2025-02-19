package types

import (
	"encoding/hex"
	"fmt"
)

type Hash [32]uint8

func (h Hash) HashToBytes() []byte {
	return h[:]
}
func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		msg := fmt.Sprintf("Invalid hash length: %d", len(b))
		panic(msg)
	}
	var bytes [32]uint8
	for i := 0; i < 32; i++ {
		bytes[i] = b[i]
	}
	return Hash(bytes)
}

func AddressFromBytes(hash [32]byte) string {
	var bytes []byte
	for index, val := range hash {
		bytes[index] = val
	}
	return hex.EncodeToString(bytes)
}

func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}
	return true
}

func (h Hash) String() string {
	return hex.EncodeToString(h.HashToBytes())
}
