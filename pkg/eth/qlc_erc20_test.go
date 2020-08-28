package eth

import (
	"context"
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
	endPoint          = "https://rinkeby.infura.io/v3/0865b420656e4d70bcbbcc76e265fd57"
	webSocketEndPoint = "wss://rinkeby.infura.io/ws/v3/0865b420656e4d70bcbbcc76e265fd57"
	mnemonic          = `lumber choice thing skull allow favorite light horse gun media treat peasant`
	contract          = "0x6d37597F0d9e917baeF2727ece52AEeb8B5294c7"
)

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
	client, err := ethclient.Dial(webSocketEndPoint)
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
