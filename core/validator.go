package core

import (
	"blockchain/pkg/e"
	"fmt"
)

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	Bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		Bc: bc,
	}
}

func (bv *BlockValidator) ValidateBlock(b *Block) error {
	if bv.Bc.HasBlock(b) {
		return e.ErrBlockKnown
	}
	if b.Height != bv.Bc.Height()+1 {
		return fmt.Errorf("invalid block height: %d, expected: %d", b.Height, bv.Bc.Height()+1)
	}

	preHeader, err := bv.Bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}
	hasher := BlockHasher{}
	prehash := hasher.Hash(preHeader)
	if prehash != b.PrevBlock {
		return fmt.Errorf("invalid prev block hash: %s, expected: %s", b.PrevBlock, prehash)
	}
	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
