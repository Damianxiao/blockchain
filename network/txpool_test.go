package network

import (
	"blockchain/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool()
	assert.Equal(t, p.Len(), 0)
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool()
	tx := core.NewTransaction([]byte("hello"))
	assert.Nil(t, p.Add(tx))
	assert.Equal(t, p.Len(), 1)
}

func TestTxPoolSort(t *testing.T) {
	p := NewTxPool()

	for i := 0; i < 1000; i++ {
		tx := core.NewTransaction([]byte("foo"))
		assert.Nil(t, p.Add(tx))
	}
	s := p.SortedTxx()
	for i := 0; i < len(s)-1; i++ {
		assert.Less(t, s[i].FirstSeen, s[i+1].FirstSeen)
	}
}
