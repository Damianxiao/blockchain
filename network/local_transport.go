package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	Addr      NetAddr
	Peers     map[NetAddr]*LocalTransport
	Lock      sync.RWMutex
	ConsumeCh chan RPC
}

func NewLocalTransport(addr NetAddr) Transport {
	return &LocalTransport{
		Addr:      addr,
		Peers:     make(map[NetAddr]*LocalTransport),
		ConsumeCh: make(chan RPC, 1024),
	}
}

func (t *LocalTransport) GetAddr() NetAddr {
	return t.Addr
}

func (t *LocalTransport) Connect(tr Transport) error {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	t.Peers[tr.GetAddr()] = tr.(*LocalTransport)

	return nil
}

func (t *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
	t.Lock.RLock()
	defer t.Lock.RUnlock()

	peer, ok := t.Peers[to]
	if !ok {
		return fmt.Errorf("%s: could not send message to %s", t.Addr, to)
	}
	peer.ConsumeCh <- RPC{
		From:    t.Addr,
		Payload: payload,
	}
	return nil
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.ConsumeCh
}

func (t *LocalTransport) Broadcast(b []byte) error {
	// ! here is t send to peers , be care
	for _, peer := range t.Peers {
		if err := t.SendMessage(peer.Addr, b); err != nil {
			return err
		}
		return nil
	}
	return nil
}
