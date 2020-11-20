package sign

import (
	"fmt"
	"math/big"
	"testing"
)

func TestNewQLCChain(t *testing.T) {
	amount := big.NewInt(189897676765000)
	userAddr := "f6933949C4096670562a5E3a21B8c29c2aacA505"
	neoTx := "1d3f2eb6d6c73b2c4ca325c8ac18141577761e43abb0154412ef4d36b11ff1b4"
	s, err := SignData(amount, userAddr, neoTx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("sig2, ", s)
}
