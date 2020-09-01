package neo

import (
	"encoding/hex"
	"math/big"
	"sort"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/core/transaction"

	"github.com/nspcc-dev/neo-go/pkg/util"

	u "github.com/qlcchain/qlc-hub/pkg/util"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

var (
	url             = "http://seed2.ngd.network:20332"
	contractAddress = "0533290f35572cd06e3667653255ffd6ee6430fb"
	contractLE, _   = util.Uint160DecodeStringLE(contractAddress)

	userWif            = "L2Dse3swNDZkwq2fkP5ctDMWB7x4kbvpkhzMJQ7oY9J2WBCATokR"
	userAccount, _     = wallet.NewAccountFromWIF(userWif)
	userAccountUint, _ = address.StringToUint160(userAccount.Address)

	wrapperWif            = "L2BAaQsPTDxGu1D9Q3x9ZS2ipabyzjBCNJAdP3D3NwZzL6KUqEkg"
	wrapperAccount, _     = wallet.NewAccountFromWIF(wrapperWif)
	wrapperAccountUint, _ = address.StringToUint160(wrapperAccount.Address)

	userEthAddress = "2e1ac6242bb084029a9eb29dfb083757d27fced4"
)

func TestNeoTransaction_QuerySwapInfo(t *testing.T) {
	c, err := NewTransaction(url, contractAddress)
	if err != nil {
		t.Fatal(err)
	}

	rHash := "6c428bbdb7b7a3c235f16d241916337d457b2e52147cb213853f1316aff2e3d3"
	r, err := c.QuerySwapInfo(rHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.ToIndentString(r))
}

func TestSortWitness(t *testing.T) {
	var scripts []transaction.Witness
	scripts = append(scripts, transaction.Witness{
		InvocationScript:   nil,
		VerificationScript: []byte{},
	})
	data, _ := hex.DecodeString("2103d0b9bfd890adc27663534b7c08131392fc67cf2e3a63e0eccc9f0a2f6d7b3e84ac")
	scripts = append(scripts, transaction.Witness{
		InvocationScript:   nil,
		VerificationScript: data,
	})
	t.Log(u.ToIndentString(scripts))
	sort.Slice(scripts, func(i, j int) bool {
		b1 := util.ArrayReverse(scripts[i].VerificationScript)
		b2 := util.ArrayReverse(scripts[j].VerificationScript)
		return big.NewInt(0).SetBytes(b1).Cmp(big.NewInt(0).SetBytes(b2)) > 0
	})
	t.Log(u.ToIndentString(scripts))
}
