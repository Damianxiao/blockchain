package network

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/types"
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
)

var DefaultBlocktime = time.Second * 5

// 代表默认node配置
type ServerOpts struct {
	Id         string
	Transports []Transport
	RPCHandler RPCHandler
	PrivateKey *crypto.PrivateKey
	BlockTime  time.Duration
	Chain      *core.Blockchain
	Logger     log.Logger
}

type Server struct {
	ServerOpts
	IsValidator bool
	BlockTime   time.Duration
	MemPool     *TxPool
	RpcCh       chan RPC
	QuitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = DefaultBlocktime
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "addr", opts.Id)
	}
	chain := core.NewBlockChain(opts.Logger, GenesisBlock())
	opts.Chain = chain

	s := &Server{
		ServerOpts:  opts,
		RpcCh:       make(chan RPC),
		QuitCh:      make(chan struct{}, 1),
		MemPool:     NewTxPool(),
		IsValidator: opts.PrivateKey != nil,
		BlockTime:   opts.BlockTime,
	}
	if opts.RPCHandler == nil {
		opts.RPCHandler = NewDefaultHandler(s)
	}
	s.RPCHandler = opts.RPCHandler
	if s.IsValidator {
		go s.ValidatorLoop()
	}
	return s
}

func (s *Server) Start() {
	s.InitServer()
free:
	for {
		select {
		case rpc := <-s.RpcCh: // 监听 RpcCh
			msg, err := s.RPCHandler.ProcessRPC(rpc)
			if err != nil {
				logrus.Error(err)
			}
			// s.Logger.Log("msg", fmt.Sprintf("tr[%s] received a msg from [%s]", s.Id, msg.From), "from", msg.From, "msg", msg)
			if err := s.ProcessMessage(msg); err != nil {
				s.Logger.Log("erraccur", err)
			}
		case <-s.QuitCh:
			break free
		}
	}
	logrus.Info("server stopped")
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
	s.Logger.Log("msg", "transaction received and added to pool", "hash", hash, "mempoolLen", s.MemPool.Len())

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
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(msg.From, t)
	case *core.Block:
		return s.ProcessBlock(t)
	default:
		return fmt.Errorf("unknown message type: %T", t)
	}

}

func (s *Server) Broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
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

func (s *Server) InitServer() {
	for _, transport := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.RpcCh <- rpc
			}
		}(transport)
	}
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
