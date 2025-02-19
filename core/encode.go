package core

import (
	"blockchain/idl/pb"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type TxEncoder struct {
	W io.Writer
}

func NewTxEncoder(w io.Writer) *TxEncoder {
	return &TxEncoder{
		W: w,
	}
}

func (enc *TxEncoder) Encode(t *Transaction) error {
	pb := t.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %s", err)
	}
	_, err = enc.W.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write tx: %s", err)
	}
	return nil
}

type TxDecoder struct {
	R io.Reader
}

func NewTxDecoder(r io.Reader) *TxDecoder {
	return &TxDecoder{
		R: r,
	}
}

func (dec *TxDecoder) Decode(t *Transaction) error {
	pb := &pb.Transaction{}
	data, err := io.ReadAll(dec.R)
	if err != nil {
		return fmt.Errorf("failed to read tx data %s", err)
	}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tx data %s", err)
	}
	*t = *TxFromProto(pb)
	return nil
}

type BlockEncoder struct {
	W io.Writer
}

func NewBlockEncoder(w io.Writer) *BlockEncoder {
	return &BlockEncoder{
		W: w,
	}
}

func (enc *BlockEncoder) Encode(b *Block) error {
	pb := b.ToProto()
	data, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %s", err)
	}
	_, err = enc.W.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write block: %s", err)
	}
	return nil
}

type BlockDecoder struct {
	R io.Reader
}

func NewBlockDecoder(r io.Reader) *BlockDecoder {
	return &BlockDecoder{
		R: r,
	}
}

func (dec *BlockDecoder) Decode(b *Block) error {
	pb := &pb.Block{}
	// 直接全部读出来 这样不需要额外判断buf中的数据大小
	data, err := io.ReadAll(dec.R)
	if err != nil {
		return fmt.Errorf("failed to read encoded block data: %w", err)
	}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		return fmt.Errorf("failed to unmarshal block data %w", err)
	}
	b.FromProto(pb)
	return nil
}

type MessageEncoder struct {
	w io.Writer
}

func NewMessageEncoder(w io.Writer) *MessageEncoder {
	return &MessageEncoder{
		w: w,
	}
}
