package network

import "net"

type NetAddr net.Addr

type Transport interface {
	Consume() <-chan RPC
	Broadcast([]byte) error
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	GetAddr() NetAddr
}
