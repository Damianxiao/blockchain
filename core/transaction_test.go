package core

import (
	"blockchain/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func RandomTxWithSignature() *Transaction {
	privKey := crypto.GenerateKeyPair()
	tx := NewTransaction([]byte("fooo"))
	tx.Sign(&privKey)
	return tx
}

func TestTxWrongData(t *testing.T) {
	pri := crypto.GenerateKeyPair()
	tx := NewTransaction([]byte("fooo"))
	assert.Nil(t, tx.Sign(&pri))
	assert.Nil(t, tx.Verify())
	tx.Data = []byte("bar")
	assert.NotNil(t, tx.Verify())
}

func TestTxWrongFrom(t *testing.T) {
	tx := RandomTxWithSignature()
	assert.Nil(t, tx.Verify())
	pri2 := crypto.GenerateKeyPair()
	pub := pri2.PublicKey()
	tx.From = pub
	assert.NotNil(t, tx.Verify())
}
