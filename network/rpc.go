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
	MessageGetStatus
	MessageStatus
	// * if anyone node is unprepare , use getblock to sync with block chain
	MessageGetBlocks
	MessageSyncBlocks
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
	// here need assert
	Data any
}

func (h *DefaultHandler) ProcessRPC(rpc RPC) (*DecodeMessage, error) {
	msg := &Message{}
	buf := bytes.NewBuffer(rpc.Payload) // 直接用 rpc.Payload 初始化 Buffer

	if err := gob.NewDecoder(buf).Decode(msg); err != nil {
		return nil, err
	}
	buf.Reset()
	buf.Write(msg.Data)

	switch msg.Header {
	case MessageTx:
		tx := core.NewTransaction([]byte{})
		if err := core.NewTxDecoder(buf).Decode(tx); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageBlock:
		block := core.NewBlock(&core.Header{}, nil)
		if err := core.NewBlockDecoder(buf).Decode(block); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageGetStatus:
		return &DecodeMessage{
			From: rpc.From,
			Data: NewGetStatusMessage(),
		}, nil
	case MessageStatus:
		status := &StatusMessage{}
		if err := gob.NewDecoder(buf).Decode(status); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: status,
		}, nil
	case MessageGetBlocks:
		return &DecodeMessage{
			From: rpc.From,
			Data: NewGetBlocksMessage(),
		}, nil
	case MessageSyncBlocks:
		syncBlocks := &SyncBlocksMessage{}
		if err := gob.NewDecoder(buf).Decode(syncBlocks); err != nil {
			return nil, err
		}
		return &DecodeMessage{
			From: rpc.From,
			Data: syncBlocks,
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
