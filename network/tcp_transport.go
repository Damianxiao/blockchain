package network

import (
	"fmt"
	"io"
	"log/slog"
	"net"
)

type TcpPeer struct {
	Conn   net.Conn
	IsDial bool
}

type TcpTransport struct {
	PeerCh     chan *TcpPeer
	ListenAddr string
	Listener   net.Listener
}

func NewTcpTransport(addr string, peerch chan *TcpPeer) *TcpTransport {
	return &TcpTransport{
		PeerCh:     peerch,
		ListenAddr: addr,
	}
}

func (peer *TcpPeer) Send(b []byte) error {
	_, err := peer.Conn.Write(b)
	return err
}

func (peer *TcpPeer) readLoop(rpcCh chan RPC) {
	readBuf := make([]byte, 2048)
	for {
		n, err := peer.Conn.Read(readBuf)
		if err != nil {
			if err == io.EOF {
				slog.Info("dial conn close", "from", peer.Conn.RemoteAddr())
			}
			slog.Error("read error", "errMsg", err, "from", peer.Conn.RemoteAddr())
			return
		}
		msg := readBuf[:n]
		// rpc
		rpcCh <- RPC{
			From:    peer.Conn.RemoteAddr(),
			Payload: msg,
		}
		// fmt.Println("read data", string(readBuf))
	}
}

func (tcp *TcpTransport) acceptLoop() {
	for {
		conn, err := tcp.Listener.Accept()
		if err != nil {

			slog.Error("accept error", "from", err)
			break
		}
		peer := &TcpPeer{
			Conn: conn}

		// slog.Info("=>new cominng connection", "from", conn.RemoteAddr(), "port", tcp.Listener.Addr())
		tcp.PeerCh <- peer
	}
}

func (tcp *TcpTransport) start() error {
	ln, err := net.Listen("tcp", tcp.ListenAddr)
	if err != nil {
		return err
	}
	fmt.Println("tcp is listening addr:", tcp.ListenAddr)
	tcp.Listener = ln
	go tcp.acceptLoop()
	select {}
}
