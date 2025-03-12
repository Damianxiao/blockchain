package network

import "blockchain/core"

type GetStatusMessage struct{}

func NewGetStatusMessage() *GetStatusMessage {
	return &GetStatusMessage{}
}

type GetBlocksMessage struct {
}

func NewGetBlocksMessage() *GetBlocksMessage {
	return &GetBlocksMessage{}
}

type SyncBlocksMessage struct {
	Blocks []*core.Block
}

type StatusMessage struct {
	Id            string
	Version       string
	CurrentHeight uint32
}

func NewStatus(id, version string, height uint32) *StatusMessage {
	return &StatusMessage{
		Id:            id,
		Version:       version,
		CurrentHeight: height,
	}
}
