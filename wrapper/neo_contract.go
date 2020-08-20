package wrapper

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
	"github.com/nspcc-dev/neo-go/pkg/interop/util"
	"time"
)

const (
	decimals   = 8
	multiplier = 100000000
)

const (
	Nep5VerifyLoopTime = 10 * time.Second
)

// Token holds all token info
type Token struct {
	// Token name
	Name string
	// Ticker symbol
	Symbol string
	// Amount of decimals
	Decimals int
	// Token owner address
	Owner []byte
	// Total tokens * multiplier
	TotalSupply int
	// Storage key for circulation value
	CirculationKey string
}

// getIntFromDB is a helper that checks for nil result of storage.Get and returns
// zero as the default value.
func getIntFromDB(ctx storage.Context, key []byte) int {
	var res int
	val := storage.Get(ctx, key)
	if val != nil {
		res = val.(int)
	}
	return res
}

// GetSupply gets the token totalSupply value from VM storage
func (t Token) GetSupply(ctx storage.Context) interface{} {
	return getIntFromDB(ctx, []byte(t.CirculationKey))
}

// BalanceOf gets the token balance of a specific address
func (t Token) BalanceOf(ctx storage.Context, holder []byte) interface{} {
	return getIntFromDB(ctx, holder)
}

// Transfer token from one user to another
func (t Token) Transfer(ctx storage.Context, from []byte, to []byte, amount int) bool {
	amountFrom := t.CanTransfer(ctx, from, to, amount)
	if amountFrom == -1 {
		return false
	}

	if amountFrom == 0 {
		storage.Delete(ctx, from)
	}

	if amountFrom > 0 {
		diff := amountFrom - amount
		storage.Put(ctx, from, diff)
	}

	amountTo := getIntFromDB(ctx, to)
	totalAmountTo := amountTo + amount
	storage.Put(ctx, to, totalAmountTo)
	runtime.Notify("transfer", from, to, amount)
	return true
}

// CanTransfer returns the amount it can transfer
func (t Token) CanTransfer(ctx storage.Context, from []byte, to []byte, amount int) int {
	if len(to) != 20 || !IsUsableAddress(from) {
		return -1
	}

	amountFrom := getIntFromDB(ctx, from)
	if amountFrom < amount {
		return -1
	}

	// Tell Transfer the result is equal - special case since it uses Delete
	if amountFrom == amount {
		return 0
	}

	// return amountFrom value back to Transfer, reduces extra Get
	return amountFrom
}

// IsUsableAddress checks if the sender is either the correct NEO address or SC address
func IsUsableAddress(addr []byte) bool {
	if len(addr) == 20 {

		if runtime.CheckWitness(addr) {
			return true
		}

		// Check if a smart contract is calling scripthash
		callingScriptHash := runtime.GetCallingScriptHash()
		if util.Equals(callingScriptHash, addr) {
			return true
		}
	}

	return false
}

// Mint initial supply of tokens.
func (t Token) Mint(ctx storage.Context, to []byte) bool {
	if !IsUsableAddress(t.Owner) {
		return false
	}
	minted := storage.Get(ctx, []byte("minted"))
	if minted != nil && minted.(bool) == true {
		return false
	}

	storage.Put(ctx, to, t.TotalSupply)
	storage.Put(ctx, []byte("minted"), true)
	runtime.Notify("transfer", "", to, t.TotalSupply)
	return true
}

var owner = util.FromAddress("NMipL5VsNoLUBUJKPKLhxaEbPQVCZnyJyB")

// createToken initializes the Token Interface for the Smart Contract to operate with
func createToken() Token {
	return Token{
		Name:           "Awesome NEO Token",
		Symbol:         "ANT",
		Decimals:       decimals,
		Owner:          owner,
		TotalSupply:    11000000 * multiplier,
		CirculationKey: "TokenCirculation",
	}
}

// // Main function = contract entry
// func Main(operation string, args []interface{}) interface{} {
// 	if operation == "name" {
// 		return Name()
// 	}
// 	if operation == "symbol" {
// 		return Symbol()
// 	}
// 	if operation == "decimals" {
// 		return Decimals()
// 	}

// 	if operation == "totalSupply" {
// 		return TotalSupply()
// 	}

// 	if operation == "balanceOf" {
// 		hodler := args[0].([]byte)
// 		return BalanceOf(hodler)
// 	}

// 	if operation == "transfer" && checkArgs(args, 3) {
// 		from := args[0].([]byte)
// 		to := args[1].([]byte)
// 		amount := args[2].(int)
// 		return Transfer(from, to, amount)
// 	}

// 	if operation == "mint" && checkArgs(args, 1) {
// 		addr := args[0].([]byte)
// 		return Mint(addr)
// 	}

// 	return true
// }

// checkArgs checks args array against a length indicator
func checkArgs(args []interface{}, length int) bool {
	if len(args) == length {
		return true
	}

	return false
}

// Name returns the token name
func Name() string {
	t := createToken()
	return t.Name
}

// Symbol returns the token symbol
func Symbol() string {
	t := createToken()
	return t.Symbol
}

// Decimals returns the token decimals
func Decimals() int {
	t := createToken()
	return t.Decimals
}

// TotalSupply returns the token total supply value
func TotalSupply() interface{} {
	t := createToken()
	ctx := storage.GetContext()
	return t.GetSupply(ctx)
}

// BalanceOf returns the amount of token on the specified address
func BalanceOf(holder []byte) interface{} {
	t := createToken()
	ctx := storage.GetContext()
	return t.BalanceOf(ctx, holder)
}

// Transfer token from one user to another
func Transfer(from []byte, to []byte, amount int) bool {
	t := createToken()
	ctx := storage.GetContext()
	return t.Transfer(ctx, from, to, amount)
}

// Mint initial supply of tokens
func Mint(to []byte) bool {
	t := createToken()
	ctx := storage.GetContext()
	return t.Mint(ctx, to)
}

func (w *WrapperServer) Nep5TransactionVerifyTry(txhash string) (status int) {
	ret := CchTransactionVerifyStatusUnknown
	return ret
}

func (w *WrapperServer) Nep5ContractWrapperLock(lockhash string) (status int) {
	ret := CchTransactionVerifyStatusUnknown
	return ret
}

//Nep5TransactionVerifyLoop tx verify loop
func (w *WrapperServer) Nep5TransactionVerifyLoop(txhash string) (status int) {
	ticker := time.NewTicker(Nep5VerifyLoopTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//verify
			ret := w.Nep5TransactionVerifyTry(txhash)
			if ret >= 0 {
				return ret
			}
		}
	}
}
