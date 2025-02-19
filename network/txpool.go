package network

import (
	"blockchain/core"
	"blockchain/types"
	"sort"
)

type TxMapSorter struct {
	txx []*core.Transaction
}

func NewTxSorter(txMap map[types.Hash]*core.Transaction) *TxMapSorter {
	txx := make([]*core.Transaction, len(txMap))
	i := 0
	for _, tx := range txMap {
		txx[i] = tx
		i++
	}
	s := &TxMapSorter{txx}
	sort.Sort(s)
	return s
}

// 获取pool排序list， 按照
func (p *TxPool) SortedTxx() []*core.Transaction {
	s := NewTxSorter(p.Transactions)
	return s.txx
}

func (s *TxMapSorter) Less(i, j int) bool {
	return s.txx[i].FirstSeen < s.txx[j].FirstSeen
}

func (s *TxMapSorter) Swap(i, j int) {
	s.txx[i], s.txx[j] = s.txx[j], s.txx[i]
}

func (s *TxMapSorter) Len() int {
	return len(s.txx)
}

type TxPool struct {
	Transactions map[types.Hash]*core.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		Transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (p *TxPool) Add(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	p.Transactions[hash] = tx
	return nil
}

func (p *TxPool) Has(hash types.Hash) bool {
	_, ok := p.Transactions[hash]
	return ok
}

func (p *TxPool) Len() int {
	return len(p.Transactions)
}

func (p *TxPool) Flush() {
	p.Transactions = make(map[types.Hash]*core.Transaction)
}

func (p *TxPool) Pending() []*core.Transaction {
	// TODO how many tx would be add?
	sortedTx := p.SortedTxx()
	p.Flush()
	return sortedTx
}
