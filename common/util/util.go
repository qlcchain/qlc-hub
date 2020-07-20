package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
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
