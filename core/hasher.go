package core

import (
	"blockchain/types"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func NewBlockHasher() BlockHasher {
	return BlockHasher{}
}

func (BlockHasher) Hash(h *Header) types.Hash {
	hash := sha256.Sum256(h.HeaderBytes())
	return types.Hash(hash)
}

type TxHasher struct {
}

func NewTxHasher() TxHasher {
	return TxHasher{}
}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Data)
	binary.Write(buf, binary.LittleEndian, tx.To)
	binary.Write(buf, binary.LittleEndian, tx.From)
	binary.Write(buf, binary.LittleEndian, tx.Value)
	binary.Write(buf, binary.LittleEndian, tx.Nonce)

	return types.Hash(sha256.Sum256(buf.Bytes()))
}
