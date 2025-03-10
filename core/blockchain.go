package core

import (
	"blockchain/pkg/e"
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	Store         Storage
	Lock          sync.RWMutex
	Headers       []*Header
	Validator     Validator
	Logger        log.Logger
	ContractState *contractState
}

func NewBlockChain(log log.Logger, genesis *Block) *Blockchain {
	bc := &Blockchain{
		Headers:       []*Header{},
		Store:         NewStorage(),
		Logger:        log,
		ContractState: NewContractState(),
	}
	bc.Validator = NewBlockValidator(bc)
	bc.AddBlockWithoutValidate(genesis)
	return bc
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.Validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.Validator.ValidateBlock(b); err != nil {
		return fmt.Errorf("block validation failed: %v", err)
	}
	// VM
	for _, tx := range b.Transaction {
		vm := NewVM(tx.Data, bc.ContractState)
		if err := vm.run(); err != nil {
			bc.Logger.Log("execute tx instructions err", err, "hash", tx.hash)
		}
		bc.Logger.Log("msg", "contract been exec", "data", bc.ContractState.data)
	}

	return bc.AddBlockWithoutValidate(b)
}

func (bc *Blockchain) Height() uint32 {

	return uint32(len(bc.Headers) - 1)
}

func (bc *Blockchain) AddBlockWithoutValidate(b *Block) error {
	bc.Headers = append(bc.Headers, b.Header)
	// logger should here
	bc.Logger.Log("msg", "new block created", "hash", NewBlockHasher().Hash(b.Header), "height", b.Height, "blockchain height", bc.Height())

	return nil

}

func (bc *Blockchain) HasBlock(b *Block) bool {
	return b.Height <= bc.Height()
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {

	if height > bc.Height() {
		return nil, e.ErrBlockUnKnown
	}
	return bc.Headers[height], nil
}
