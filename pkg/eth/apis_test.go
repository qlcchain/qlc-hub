package eth

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

const (
	endPoint   = "https://rinkeby.infura.io/v3/0865b420656e4d70bcbbcc76e265fd57"
	endPointws = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	mnemonic   = `lumber choice thing skull allow favorite light horse gun media treat peasant`
	contract   = "0x6d37597F0d9e917baeF2727ece52AEeb8B5294c7"

	wrapperPrikey = "67652fa52357b65255ac38d0ef8997b5608527a7c1d911ecefb8bc184d74e92e"
	userEthPrikey = "b44980807202aff0707cc4eebad4f9e47b4d645cf9f4320653ff62dcd5751234"
)

func TestNewQLCChain(t *testing.T) {
	client, err := ethclient.Dial(endPointws)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	//instance, opts, err := GetTransactor(client, wrapperPrikey, contract)
	//if err != nil {
	//	t.Fatal(err)
	//}
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	tx, err := WrapperLock(rHash, wrapperPrikey, contract, 100000000, client)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Wrapper Lock: ", tx)

	b, i, err := TxVerifyAndConfirmed(tx, 0, client)
	if !b || err != nil {
		t.Fatal(b, i, err)
	}
	//time.Sleep(30 * time.Second)

	tx2, err := UserUnlock(rHash, rOrigin, userEthPrikey, contract, client)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("User Unlock: ", tx2)
}

func TestNewQLCChain2(t *testing.T) {
	client, err := ethclient.Dial(endPointws)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	instance, err := GetTransactorSession(client, wrapperPrikey, contract)
	if err != nil {
		t.Fatal(err)
	}
	rOrigin, rHash := util.Sha256Hash()
	fmt.Println("hash: ", rOrigin, "==>", rHash)

	bigAmount := big.NewInt(12 * 100000000)
	rHashBytes, err := util.HexStringToBytes32(rHash)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := instance.IssueLock(rHashBytes, bigAmount)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tx.Hash().Hex())
}

func TestGetHashTimer(t *testing.T) {
	client, err := ethclient.Dial(endPointws)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	r, err := GetHashTimer(client, contract, "3315e92b49957eeeb75cdb1e57560b00ca0b2ec1240d2af194cef580ca188a02")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(util.ToIndentString(r))
}

func TestQLCChainTransactorSession_IssueLock(t *testing.T) {
	t.Skip()

	// create account from mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		t.Fatal(err)
	}

	account, err := wallet.Derive(accounts.DefaultBaseDerivationPath, false)
	if err != nil {
		t.Fatal(err)
	}

	address := account.Address
	t.Log(address.Hex())

	// create account from privatekey
	//privateKey, err := crypto.HexToECDSA("private key hex string")
	//if err != nil {
	//	t.Fatal(err)
	//}
	client, err := ethclient.Dial(endPointws)
	if err != nil {
		t.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	if privateKey, err := wallet.PrivateKey(account); err != nil {
		t.Fatal(err)
	} else {
		ctx := context.Background()
		nonce, err := client.PendingNonceAt(ctx, address)
		if err != nil {
			t.Fatal(err)
		}
		price, err := client.SuggestGasPrice(ctx)
		if err != nil {
			t.Fatal(err)
		}

		auth := bind.NewKeyedTransactor(privateKey)
		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = big.NewInt(0)      // Funds to transfer along the transaction (nil = 0 = no funds)
		auth.GasLimit = uint64(1000000) // Gas limit to set for the transaction execution (0 = estimate)
		auth.GasPrice = price           //Gas price to use for the transaction execution (nil = gas price oracle)

		contractAddress := common.HexToAddress(contract)

		// Transacting with an Ethereum contract
		transactor, err := NewQLCChainTransactor(contractAddress, client)
		if err != nil {
			t.Fatal(err)
		}

		// TODO: fill hash
		rHash := [32]byte{}
		if tx, err := transactor.IssueLock(auth, rHash, big.NewInt(100)); err == nil {
			t.Log(util.ToIndentString(tx))
		} else {
			t.Fatal(err)
		}
		c, err := NewQLCChain(contractAddress, client)
		if err != nil {
			t.Fatal(err)
		}

		// Pre-configured contract sessions
		session := &QLCChainSession{
			Contract: c,
			CallOpts: bind.CallOpts{
				Pending: true, //Whether to operate on the pending state or the last known one
			},
			TransactOpts: *auth,
		}

		if tx, err := session.IssueLock(rHash, big.NewInt(100)); err == nil {
			t.Log(util.ToIndentString(tx))
		} else {
			t.Fatal(err)
		}

		// query log
		filter, err := NewQLCChainFilterer(contractAddress, client)
		if err != nil {
			t.Fatal(err)
		}
		sink := make(chan<- *QLCChainLockedState, 1)
		if state, err := filter.WatchLockedState(&bind.WatchOpts{}, sink, [][32]byte{rHash}); err == nil {
			t.Log(util.ToIndentString(state))
		} else {
			t.Fatal(err)
		}

		// subscribe logs
		logs := make(chan<- types.Log, 10)
		if _, err := client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
		}, logs); err != nil {
			t.Fatal(err)
		}

		// for logs
	}
}
