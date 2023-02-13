package aliyundrive

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"math/big"

	"github.com/dustinxie/ecc"
)

func NewPrivateKey() (*ecdsa.PrivateKey, error) {
	p256k1 := ecc.P256k1()
	return ecdsa.GenerateKey(p256k1, rand.Reader)
}

func NewPrivateKeyFromHex(hex_ string) (*ecdsa.PrivateKey, error) {
	data, err := hex.DecodeString(hex_)
	if err != nil {
		return nil, err
	}
	return NewPrivateKeyFromBytes(data), nil

}

func NewPrivateKeyFromBytes(priv []byte) *ecdsa.PrivateKey {
	p256k1 := ecc.P256k1()
	x, y := p256k1.ScalarBaseMult(priv)
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: p256k1,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(priv),
	}
}

func PrivateKeyToHex(private *ecdsa.PrivateKey) string {
	return hex.EncodeToString(PrivateKeyToBytes(private))
}

func PrivateKeyToBytes(private *ecdsa.PrivateKey) []byte {
	return private.D.Bytes()
}

func PublicKeyToHex(public *ecdsa.PublicKey) string {
	return hex.EncodeToString(PublicKeyToBytes(public))
}

func PublicKeyToBytes(public *ecdsa.PublicKey) []byte {
	x := public.X.Bytes()
	if len(x) < 32 {
		for i := 0; i < 32-len(x); i++ {
			x = append([]byte{0}, x...)
		}
	}

	y := public.Y.Bytes()
	if len(y) < 32 {
		for i := 0; i < 32-len(y); i++ {
			y = append([]byte{0}, y...)
		}
	}
	return append(x, y...)
}
