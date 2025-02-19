package network

import (
	"blockchain/core"
	"bytes"
	"encoding/gob"
	"fmt"
)

const (
	MessageTypeiota = iota
	MessageTx
	MessageBlock
	// * if anyone node is unprepare , use getblock to sync with block chain
	MessageGetBlock
)

type RPC struct {
	From    NetAddr
	Payload []byte
}

type RPCHandler interface {
	ProcessRPC(rpc RPC) (*DecodeMessage, error)
}

type RPCProcessor interface {
	ProcessTransaction(NetAddr, *core.Transaction) error
}

func NewMessage(t int, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

type Message struct {
	Header int
	Data   []byte
}

type DefaultHandler struct {
	p RPCProcessor
}

func NewDefaultHandler(p RPCProcessor) *DefaultHandler {
	return &DefaultHandler{
		p: p,
	}
}

type DecodeMessage struct {
	From NetAddr
	Data any
}

func (h *DefaultHandler) ProcessRPC(rpc RPC) (*DecodeMessage, error) {
	msg := &Message{}
	buf := &bytes.Buffer{}
	buf.Write(rpc.Payload)
	if err := gob.NewDecoder(buf).Decode(msg); err != nil {
		return nil, err
	}
	switch msg.Header {
	case MessageTx:
		pb := new(bytes.Buffer)
		pb.Write(msg.Data)
		tx := core.NewTransaction([]byte{})
		if err := core.NewTxDecoder(pb).Decode(tx); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageBlock:
		pb := new(bytes.Buffer)
		pb.Write(msg.Data)
		block := core.NewBlock(&core.Header{}, nil)
		if err := core.NewBlockDecoder(pb).Decode(block); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: block,
		}, nil
	// TODO other case tx msg block....
	default:
		return nil, fmt.Errorf("invalid message header %v", msg.Header)
	}

}

func (m *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(m)
	return buf.Bytes()
}
