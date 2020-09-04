package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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

func NewBase64() string {
	key, _ := New()
	encode := Encode(key)
	return base64.StdEncoding.EncodeToString([]byte(encode))
}

func FromBase64(key string) (*ecdsa.PrivateKey, error) {
	if data, err := base64.StdEncoding.DecodeString(key); err != nil {
		return nil, err
	} else {
		key = string(data)
		return Decode(string(data)), nil
	}
}
