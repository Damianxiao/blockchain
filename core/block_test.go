package core

import (
	"blockchain/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockSign(t *testing.T) {
	block, _ := RandomBlock(0)
	pri := crypto.GenerateKeyPair()
	err := block.Sign(pri)
	assert.Nil(t, err)
	err = block.Verify()
	assert.Nil(t, err)

	pri2 := crypto.GenerateKeyPair()
	block.Validator = pri2.PublicKey()
	assert.NotNil(t, block.Verify())
	datahash, err := CalculateDatahash(block.Transaction)
	block.Header.DataHash = datahash
	assert.Nil(t, err)
	assert.Equal(t, datahash, block.DataHash)
}
