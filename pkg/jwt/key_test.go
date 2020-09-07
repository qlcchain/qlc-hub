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

func TestNewBase58(t *testing.T) {
	base58 := NewBase58Key()
	t.Log(base58)
	if from, err := FromBase58(base58); err != nil {
		t.Fatal(err)
	} else {
		t.Log(Encode(from))
	}
}
