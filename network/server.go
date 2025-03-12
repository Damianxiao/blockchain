package network

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/types"
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
)

var DefaultBlocktime = time.Second * 5

type ServerOpts struct {
	ListenAddress string
	NodeSeeds     []string
	RPCHandler    RPCHandler
	PrivateKey    *crypto.PrivateKey
	BlockTime     time.Duration
	Logger        log.Logger
}

type Server struct {
	ServerOpts
	TcpTransport *TcpTransport
	PeerCh       chan *TcpPeer
	PeerMap      map[NetAddr]*TcpPeer
	IsValidator  bool
	BlockTime    time.Duration
	Chain        *core.Blockchain
	MemPool      *TxPool
	RpcCh        chan RPC
	QuitCh       chan struct{}
	mu           sync.RWMutex
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = DefaultBlocktime
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "node", opts.ListenAddress)
	}
	chain := core.NewBlockChain(opts.Logger, GenesisBlock())

	s := &Server{
		PeerMap:     make(map[NetAddr]*TcpPeer),
		ServerOpts:  opts,
		RpcCh:       make(chan RPC),
		QuitCh:      make(chan struct{}, 1),
		MemPool:     NewTxPool(),
		IsValidator: opts.PrivateKey != nil,
		BlockTime:   opts.BlockTime,
	}
	s.Chain = chain
	peerch := make(chan *TcpPeer)
	// new a tcp transport
	s.TcpTransport = NewTcpTransport(opts.ListenAddress, peerch)
	s.PeerCh = s.TcpTransport.PeerCh
	if opts.RPCHandler == nil {
		opts.RPCHandler = NewDefaultHandler(s)
	}
	s.RPCHandler = opts.RPCHandler

	if s.IsValidator {
		go s.ValidatorLoop()
	}

	// get now blockchain state
	return s
}

// connect each other
func (s *Server) connectToNodeFromSeeds() {
	for _, netaddr := range s.NodeSeeds {
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				s.Logger.Log("Seeds initialize err", err)
				return
			}
			peer := &TcpPeer{
				Conn:   conn,
				IsDial: true,
			}
			s.PeerCh <- peer
		}(netaddr)
	}
}

func (s *Server) Start() {
	go s.TcpTransport.start()
	time.Sleep(1 * time.Second)

	s.connectToNodeFromSeeds()
	time.Sleep(1 * time.Second)
free:
	for {
		select {
		case peer := <-s.PeerCh:
			s.Logger.Log("=> new peer from", peer.Conn.RemoteAddr())
			go peer.readLoop(s.RpcCh)
			s.mu.Lock()
			s.PeerMap[peer.Conn.RemoteAddr()] = peer
			s.mu.Unlock()
			err := s.sendGetStatusMessage(peer)
			if err != nil {
				s.Logger.Log("sync request send fail", err)
			}
		case rpc := <-s.RpcCh:
			// s.Logger.Log("received rpc from:", rpc.From)
			msg, err := s.RPCHandler.ProcessRPC(rpc)
			if err != nil {
				logrus.Error(err)
			}
			if err := s.ProcessMessage(msg); err != nil {
				s.Logger.Log("[ProcessMessage]err", err, "msg", fmt.Sprintf("%v", msg))
			}
		case <-s.QuitCh:
			break free
		}
	}
	s.Logger.Log("msg", "server stopped")
}

func (s *Server) sendGetStatusMessage(peer *TcpPeer) error {
	var (
		header           = MessageGetStatus
		getStatusMessage = NewGetStatusMessage()
	)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getStatusMessage); err != nil {
		return err
	}
	msg := NewMessage(header, buf.Bytes())
	//  for test
	return peer.Send(msg.Bytes())
}

func (s *Server) ProcessGetStatus(from NetAddr, msg *GetStatusMessage) error {
	buf := new(bytes.Buffer)
	status := NewStatus(s.ListenAddress, "version:0.0.1", s.Chain.Height())
	if err := gob.NewEncoder(buf).Encode(status); err != nil {
		return err
	}
	newMessage := NewMessage(MessageStatus, buf.Bytes())
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, ok := s.PeerMap[from]
	if !ok {
		return fmt.Errorf("send node doesnt exist")
	}
	s.Logger.Log("msg", "sent status to", "to", from, "status", fmt.Sprintf("%v", status))
	return peer.Send(newMessage.Bytes())
}

func (s *Server) ProcessStatus(from NetAddr, msg *StatusMessage) error {
	s.Logger.Log("msg", "received status", "status", fmt.Sprintf("%v", msg))

	if s.Chain.Height() >= msg.CurrentHeight {
		s.Logger.Log("msg", "cant sync the chain below slef", "currentHeight", s.Chain.Height(), "but:", msg.CurrentHeight)
		return nil
	}
	getBlockMessage := GetBlocksMessage{}
	peer, ok := s.PeerMap[from]
	if !ok {
		return fmt.Errorf("send node doesnt exist")

	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlockMessage); err != nil {
		return err
	}
	NewMessage := NewMessage(MessageGetBlocks, buf.Bytes())
	s.Logger.Log("msg", "send block sync request!!!!!", "currentHeight", s.Chain.Height(), "but:", msg.CurrentHeight)
	return peer.Send(NewMessage.Bytes())
}

func (s *Server) ProcessGetBlock(from NetAddr, msg *GetBlocksMessage) error {
	// buf := new(bytes.Buffer)
	blocksMessage := SyncBlocksMessage{}
	for i := 1; i <= int(s.Chain.Height()); i++ {
		block, err := s.Chain.GetBlock(uint32(i))
		if err != nil {
			return err
		}
		blocksMessage.Blocks = append(blocksMessage.Blocks, block)
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMessage); err != nil {
		return err
	}
	NewMesage := NewMessage(MessageSyncBlocks, buf.Bytes())
	peer := s.PeerMap[from]
	return peer.Send(NewMesage.Bytes())
}

func (s *Server) ProcessSyncBlocks(msg *SyncBlocksMessage) error {
	s.Logger.Log("msg", "received sync block!", "blocks", msg.Blocks, "len", len(msg.Blocks))
	var oldHeight = s.Chain.Height()
	for _, block := range msg.Blocks {
		err := s.Chain.AddBlockWithoutValidate(block)
		if err != nil {
			return err
		}
	}
	s.Logger.Log("msg", "sync block success!", "syncStatus:", fmt.Sprintf("oldHeight%v => nowHeight%v", oldHeight, s.Chain.Height()))
	return nil
}

func (s *Server) ProcessTransaction(from NetAddr, tx *core.Transaction) error {
	hash := tx.Hash(core.NewTxHasher())

	if s.MemPool.Has(hash) {
		// TODO
		return nil
	}
	if err := tx.Verify(); err != nil {
		return err
	}
	tx.FirstSeen = time.Now().UnixNano()
	s.Logger.Log("msg", "transaction received and added to pool", "from", from, "hash", hash, "mempoolLen", s.MemPool.Len())

	//  broadcast tx
	go s.BroadcastTx(tx)

	return s.MemPool.Add(tx)
}

func (s *Server) ProcessBlock(b *core.Block) error {
	if err := s.Chain.AddBlock(b); err != nil {
		return err
	}
	s.Logger.Log("msg", "received a new block and added", "height", b.Header.Height, "hash", core.NewBlockHasher().Hash(b.Header))
	// * if block is valid , broadcast it
	go s.BroadcastBlock(b)
	return nil
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {
	// here is origin msg decode
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(msg.From, t)
	case *core.Block:
		return s.ProcessBlock(t)
	case *GetStatusMessage:
		return s.ProcessGetStatus(msg.From, t)
	case *StatusMessage:
		return s.ProcessStatus(msg.From, t)
	case *GetBlocksMessage:
		return s.ProcessGetBlock(msg.From, t)
	case *SyncBlocksMessage:
		return s.ProcessSyncBlocks(t)
	default:
		return fmt.Errorf("unknown message type: %T", t)
	}

}

func (s *Server) Broadcast(payload []byte) error {
	for addr, peer := range s.PeerMap {
		if err := peer.Send(payload); err != nil {
			s.Logger.Log("msg", "failed to broadcast to peer", "addr", addr, "err", err)
		}
	}
	return nil
}

func (s *Server) ValidatorLoop() error {
	ticker := time.NewTicker(s.BlockTime)
	for {
		<-ticker.C
		err := s.CreateBlock()
		if err != nil {
			logrus.Error(err)
		}
	}
}
func (s *Server) BroadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := core.NewTxEncoder(buf).Encode(tx); err != nil {
		return err
	}
	msg := NewMessage(MessageTx, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) BroadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	err := core.NewBlockEncoder(buf).Encode(b)
	if err != nil {
		return err
	}
	msg := NewMessage(MessageBlock, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) CreateBlock() error {
	header, err := s.Chain.GetHeader(s.Chain.Height())
	if err != nil {
		return err
	}

	txx := s.MemPool.Pending()
	newBlock, err := core.NewBLockFromHeader(header, txx)
	if err != nil {
		return err
	}
	// sign
	if err := newBlock.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.Chain.AddBlock(newBlock); err != nil {
		return err
	}
	// validator broadcast
	go s.BroadcastBlock(newBlock)
	return nil
}

func GenesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		Height:    0,
		DataHash:  types.Hash{},
		TimeStamp: 0000000,
	}
	return core.NewBlock(header, nil)
}
