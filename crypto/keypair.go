package crypto

import (
	"blockchain/idl/pb"
	"blockchain/types"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"
)

type PublicKey []byte

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func NewPrivateKeyFromReader(r io.Reader) PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		panic(err)
	}
	return PrivateKey{
		key: key,
	}
}

func GenerateKeyPair() PrivateKey {
	return NewPrivateKeyFromReader(rand.Reader)
}

// compressed pubkey to 33 bytes
func (p PrivateKey) PublicKey() PublicKey {
	return elliptic.MarshalCompressed(p.key.Curve, p.key.X, p.key.Y)
}

func (p PrivateKey) Address() string {
	hash := sha256.Sum256(p.PublicKey())
	str := types.AddressFromBytes(hash)
	return str[len(str)-20:]
}

func (p PublicKey) String() string {
	return hex.EncodeToString(p)
}

type Signature struct {
	S *big.Int
	R *big.Int
}

func NewSignature() *Signature {
	return &Signature{
		S: &big.Int{},
		R: &big.Int{},
	}
}

func (p PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, p.key, data)
	if err != nil {
		panic(err)
	}
	return &Signature{
		R: r,
		S: s,
	}, nil
}

func (sig Signature) Verify(data []byte, pub PublicKey) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pub)
	pk := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return ecdsa.Verify(pk, data, sig.R, sig.S)
}

func (sig *Signature) ToProto() *pb.Signature {
	if sig.R != nil && sig.S != nil {
		return &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		}
	}
	return &pb.Signature{}
}

func FromProto(proto *pb.Signature) *Signature {
	return &Signature{
		R: new(big.Int).SetBytes(proto.R),
		S: new(big.Int).SetBytes(proto.S),
	}
}
