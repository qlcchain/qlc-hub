package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
)

//CreateDirIfNotExist create given folder
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		return err
	}
	return nil
}

func ToIndentString(v interface{}) string {
	b, err := json.MarshalIndent(&v, "", "\t")
	if err != nil {
		return ""
	}
	return string(b)
}

// Bytes fills the given byte slice with random bytes.
func Bytes(data []byte) error {
	_, err := rand.Read(data)
	return err
}

func RandomHexString(length int) string {
	if length == 0 {
		return ""
	}
	b := make([]byte, length)
	_ = Bytes(b)
	s := hex.EncodeToString(b)
	return s
}

func HexStringToBytes32(str string) ([32]byte, error) {
	if len(str) != 64 {
		return [32]byte{}, fmt.Errorf("hex str %s length %d is not right", str, len(str))
	}
	var bs [32]byte
	lock, err := hex.DecodeString(str)
	if err != nil {
		return [32]byte{}, err
	}
	copy(bs[:], lock)
	return bs, nil
}

func StringToBytes32(str string) ([32]byte, error) {
	if len(str) != 32 {
		return [32]byte{}, fmt.Errorf("str %s length %d is not right", str, len(str))
	}
	strBytes := []byte(str)
	var bs [32]byte
	copy(bs[:], strBytes)
	return bs, nil
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func Sha256Hash() (string, string) {
	rOrigin := String(32)
	h := sha256.Sum256([]byte(rOrigin))
	rHash := hex.EncodeToString(h[:])
	return rOrigin, rHash
}

func RemoveHexPrefix(str string) string {
	if strings.HasPrefix(str, "0x") {
		s := strings.TrimLeft(str, "0x")
		return s
	}
	return str
}

func IsvalidNEOAddress(addr string) bool {
	_, err := address.StringToUint160(addr)
	return err == nil
}
