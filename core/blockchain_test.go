package core

import (
	"blockchain/types"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestBlockchain(t *testing.T) {
	genesis, _ := RandomBlock(0)
	bc := NewBlockChain(log.NewLogfmtLogger(os.Stderr), genesis)

	assert.NotNil(t, bc)
	for i := 0; i < 1000; i++ {
		block, _ := RandomBlock(i + 1)
		bc.AddBlock(block)
	}
	assert.Equal(t, uint32(len(bc.Headers))-1, bc.Height())
	block1, _ := RandomBlock(int(bc.Height()))
	assert.NotNil(t, bc.AddBlock(block1))
	block2, _ := RandomBlock(int(bc.Height()) + 10)
	assert.NotNil(t, bc.AddBlock(block2))
	block3, _ := RandomBlock(int(bc.Height()))
	block3.PrevBlock = types.Hash{}
	assert.NotNil(t, block3)
}

func TestGetHeader(t *testing.T) {
	genesis, _ := RandomBlock(0)
	bc := NewBlockChain(log.NewLogfmtLogger(os.Stderr), genesis)
	assert.NotNil(t, bc)
	for i := 0; i < 1000; i++ {
		block, _ := RandomBlock(i + 1)
		bc.AddBlock(block)
	}
	header, err := bc.GetHeader(0)
	assert.Nil(t, err)
	assert.Equal(t, header.Height, uint32(0))
}
