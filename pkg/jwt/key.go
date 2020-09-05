package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"github.com/mr-tron/base58"
)

func Decode(pemEncoded string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return privateKey
}

func Encode(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

func New() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey
}

func NewBase58Key() string {
	key, _ := New()
	encode := Encode(key)

	return base58.Encode([]byte(encode))
}

func FromBase58(key string) (*ecdsa.PrivateKey, error) {
	if data, err := base58.Decode(key); err != nil {
		return nil, err
	} else {
		return Decode(string(data)), nil
	}
}
