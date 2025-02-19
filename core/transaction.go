package core

import (
	"blockchain/crypto"
	"blockchain/idl/pb"
	"blockchain/types"
	"fmt"
	"math/rand"
	"time"
)

type Transaction struct {
	Data  []byte
	To    crypto.PublicKey
	From  crypto.PublicKey
	Value uint64
	Nonce uint64

	Signature *crypto.Signature
	hash      types.Hash
	FirstSeen int64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:      data,
		Nonce:     rand.Uint64(),
		Signature: crypto.NewSignature(),
		FirstSeen: time.Now().UnixNano(),
	}
}

func (t *Transaction) Sign(pri *crypto.PrivateKey) error {
	if t.From == nil {
		t.From = pri.PublicKey()
	}
	hash := t.Hash(TxHasher{})
	// 交易的签名对象是特定的数据哈希
	sig, err := pri.Sign(hash.HashToBytes())
	if err != nil {
		return err
	}
	//交易发起方签名交易
	t.Signature = sig
	return nil
}

func (t *Transaction) Verify() error {
	if t.Signature == nil {
		return fmt.Errorf("tx signature is not exist ")
	}
	hash := t.Hash(TxHasher{})
	if ok := t.Signature.Verify(hash.HashToBytes(), t.From); !ok {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (t *Transaction) ToProto() *pb.Transaction {
	return &pb.Transaction{
		Data:      t.Data,
		To:        &pb.PublicKey{Key: t.To},
		From:      &pb.PublicKey{Key: t.From},
		Value:     t.Value,
		Nonce:     t.Nonce,
		Signature: t.Signature.ToProto(),
		FirstSeen: t.FirstSeen,
		Hash:      t.hash[:],
	}
}

func TxFromProto(proto *pb.Transaction) *Transaction {
	t := &Transaction{
		Data:      proto.Data,
		To:        proto.To.Key,
		From:      proto.From.Key,
		Value:     proto.Value,
		Nonce:     proto.Nonce,
		Signature: crypto.FromProto(proto.Signature),
		FirstSeen: proto.FirstSeen,
		hash:      types.Hash(proto.Hash),
	}
	return t
}

func (tx *Transaction) Hash(hasher TxHasher) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return hasher.Hash(tx)
}

func (tx *Transaction) SetFirstSeen(t int64) {
	tx.FirstSeen = t
}
