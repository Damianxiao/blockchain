package network

type NetAddr string

type Transport interface {
	Consume() <-chan RPC
	Broadcast([]byte) error
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	GetAddr() NetAddr
}
