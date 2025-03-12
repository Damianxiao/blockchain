package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func RandomBlock(height int) (*Block, error) {
	// 生成随机版本号
	mathrand.Seed(time.Now().UnixNano())

	// 生成随机版本号
	version := uint32(mathrand.Int31n(1000))

	// 生成随机PrevBlock
	var prevBlock types.Hash
	if _, err := rand.Read(prevBlock[:]); err != nil {
		return nil, err
	}

	// 生成随机TimeStamp
	timeStamp := time.Now().Unix()

	// 生成随机Nonce
	nonce := uint32(mathrand.Int31n(1000000))

	// 生成随机Header
	header := &Header{
		Version:   version,
		PrevBlock: prevBlock,
		TimeStamp: timeStamp,
		Nonce:     nonce,
		Height:    uint32(height),
	}

	// 生成随机Transactions
	// numTransactions := mathrand.Intn(50) // 随机生成交易数量
	numTransactions := mathrand.Intn(50) // 随机生成交易数量
	transactions := make([]*Transaction, 0)
	for i := 0; i < numTransactions; i++ {
		tx := RandomTxWithSignature()
		if tx != nil {
			transactions = append(transactions, tx)
		}
	}

	// 生成随机Validator
	var validatorPubKey crypto.PublicKey
	if _, err := rand.Read(validatorPubKey); err != nil {
		return nil, err
	}

	// 生成随机Signature
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, prevBlock[:])
	if err != nil {
		return nil, err
	}
	blockSignature := &crypto.Signature{R: r, S: s}

	// 生成随机区块哈希
	var blockHash types.Hash
	if _, err := rand.Read(blockHash[:]); err != nil {
		return nil, err
	}

	// 创建Block
	block := &Block{
		Header:      header,
		Transaction: transactions,
		Validator:   validatorPubKey,
		Signature:   blockSignature,
		hash:        blockHash,
	}

	return block, nil
}

func TestEncode(t *testing.T) {
	block, _ := RandomBlock(0)
	buf := &bytes.Buffer{}
	err := block.Encode(NewBlockEncoder(buf))
	assert.Nil(t, err)
	block2 := &Block{
		Header: &Header{},
	}
	err = block2.Decode(NewBlockDecoder(buf))
	assert.Nil(t, err)
	assert.Equal(t, block, block2)
}

func TestEncodeBlocks(t *testing.T) {
	bs := make([][]byte, 0)
	blocks := make([]*Block, 0)
	for i := 0; i < 10; i++ {
		block, _ := RandomBlock(0)
		buf := &bytes.Buffer{}
		err := block.Encode(NewBlockEncoder(buf))
		bs = append(bs, buf.Bytes())
		assert.Nil(t, err)
		blocks = append(blocks, block)
	}
	blocks2 := make([]*Block, 0)
	buf := &bytes.Buffer{}

	for _, bytes := range bs {
		buf.Write(bytes)
		block := &Block{
			Header: &Header{},
		}
		err := block.Decode(NewBlockDecoder(buf))
		assert.Nil(t, err)
		blocks2 = append(blocks2, block)
	}
	assert.Equal(t, blocks, blocks2)
}
