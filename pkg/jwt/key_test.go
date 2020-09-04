package jwt

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	priv, pub := New()
	privString := Encode(priv)
	fmt.Println(privString)
	priv2 := Decode(privString)
	if !reflect.DeepEqual(priv, priv2) {
		t.Fatal("invalid priv key")
	}
	pub2 := &priv2.PublicKey
	if !reflect.DeepEqual(pub, pub2) {
		t.Fatal("invalid pub key")
	}
}

func TestNewBase64(t *testing.T) {
	t.Log(NewBase64())
}
