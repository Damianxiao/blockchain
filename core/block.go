package core

import (
	"blockchain/crypto"
	"blockchain/idl/pb"
	"blockchain/types"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Header struct {
	Version   uint32
	PrevBlock types.Hash
	DataHash  types.Hash
	TimeStamp int64
	Nonce     uint32
	Height    uint32
}

type Block struct {
	*Header
	Transaction []*Transaction
	Validator   crypto.PublicKey
	Signature   *crypto.Signature
	hash        types.Hash
}

func NewBlock(h *Header, txx []*Transaction) *Block {
	return &Block{
		Header:      h,
		Transaction: txx,
	}
}

func (b *Block) Sign(pri crypto.PrivateKey) error {
	sig, err := pri.Sign(b.HeaderBytes())
	if err != nil {
		return fmt.Errorf("Sign block failed %s", err)
	}
	b.Signature = sig
	b.Validator = pri.PublicKey()
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block signature is nil")
	}
	sig := b.Signature
	ok := sig.Verify(b.HeaderBytes(), b.Validator)
	if !ok {
		return fmt.Errorf("verify sig fail")
	}

	//  verify tx
	for _, tx := range b.Transaction {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	dataHash, err := CalculateDatahash(b.Transaction)
	if err != nil {
		return err
	}
	if dataHash != b.DataHash {
		fmt.Errorf("block (%s) has invalid datahash", b.hash)
	}
	return nil

}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func (h *Header) HeaderBytes() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	enc.Encode(h)
	return buf.Bytes()
}

func (b *Block) ToProto() *pb.Block {
	// 预分配内存
	txx := make([]*pb.Transaction, 0, len(b.Transaction))
	for _, val := range b.Transaction {
		if val != nil {
			txx = append(txx, val.ToProto())
		}
	}
	return &pb.Block{
		Header: &pb.Header{
			Version:   b.Version,
			PrevBlock: b.PrevBlock[:],
			Datahash:  b.DataHash[:],
			Timestamp: b.TimeStamp,
			Nonce:     b.Nonce,
			Height:    b.Height,
		},
		Validator: &pb.PublicKey{
			Key: b.Validator,
		},
		Signature:    b.Signature.ToProto(),
		Hash:         b.hash[:],
		Transactions: txx,
	}
}

func (b *Block) FromProto(proto *pb.Block) {
	b.Version = proto.Header.Version
	b.PrevBlock = types.Hash(proto.Header.PrevBlock)
	b.DataHash = types.Hash(proto.Header.Datahash)
	b.TimeStamp = proto.Header.Timestamp
	b.Nonce = proto.Header.Nonce
	b.Height = proto.Header.Height
	b.Validator = proto.Validator.Key
	b.Signature = crypto.FromProto(proto.Signature)
	b.hash = types.Hash(proto.Hash)
	b.Transaction = make([]*Transaction, 0, len(proto.Transactions))

	for _, txProto := range proto.Transactions {
		tx := TxFromProto(txProto)
		b.Transaction = append(b.Transaction, tx)
	}
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func CalculateDatahash(txx []*Transaction) (types.Hash, error) {
	buf := &bytes.Buffer{}
	for _, tx := range txx {
		if err := NewTxEncoder(buf).Encode(tx); err != nil {
			return types.Hash{}, err
		}
	}
	hash := sha256.Sum256(buf.Bytes())
	return hash, nil
}

func NewBLockFromHeader(h *Header, txx []*Transaction) (*Block, error) {
	preHash := NewBlockHasher().Hash(h)
	datahash, err := CalculateDatahash(txx)
	if err != nil {
		return nil, err
	}
	newBlock := &Block{
		Header: &Header{
			Version:   h.Version,
			Height:    h.Height + 1,
			PrevBlock: preHash,
			TimeStamp: time.Now().UnixNano(),
			DataHash:  datahash,
		},
		Transaction: txx,
	}
	return newBlock, nil
}
