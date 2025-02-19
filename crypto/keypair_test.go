package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeypairFail(t *testing.T) {
	pri := GenerateKeyPair()
	// pub := pri.PublicKey()
	data := []byte("hello")
	sig, err := pri.Sign(data)
	assert.Nil(t, err)
	otherPri := GenerateKeyPair()
	otherPub := otherPri.PublicKey()
	ok := sig.Verify(data, otherPub)
	assert.False(t, ok)

}

func TestKeypairSuccess(t *testing.T) {
	pri := GenerateKeyPair()
	pub := pri.PublicKey()
	data := []byte("hello world!")
	sig, err := pri.Sign(data)
	assert.Nil(t, err)
	ok := sig.Verify(data, pub)
	assert.True(t, ok)
}
