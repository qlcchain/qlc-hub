// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package eth

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// QLCChainABI is the input ABI used to generate the binding from.
const QLCChainABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"state\",\"type\":\"uint256\"}],\"name\":\"LockedState\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"issueLock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rOrigin\",\"type\":\"bytes32\"}],\"name\":\"issueUnlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"}],\"name\":\"issueFetch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"executor\",\"type\":\"address\"}],\"name\":\"destoryLock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rOrigin\",\"type\":\"bytes32\"}],\"name\":\"destoryUnlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"}],\"name\":\"destoryFetch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rHash\",\"type\":\"bytes32\"}],\"name\":\"hashTimer\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// QLCChain is an auto generated Go binding around an Ethereum contract.
type QLCChain struct {
	QLCChainCaller     // Read-only binding to the contract
	QLCChainTransactor // Write-only binding to the contract
	QLCChainFilterer   // Log filterer for contract events
}

// QLCChainCaller is an auto generated read-only Go binding around an Ethereum contract.
type QLCChainCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QLCChainTransactor is an auto generated write-only Go binding around an Ethereum contract.
type QLCChainTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QLCChainFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type QLCChainFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QLCChainSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type QLCChainSession struct {
	Contract     *QLCChain         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// QLCChainCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type QLCChainCallerSession struct {
	Contract *QLCChainCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// QLCChainTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type QLCChainTransactorSession struct {
	Contract     *QLCChainTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// QLCChainRaw is an auto generated low-level Go binding around an Ethereum contract.
type QLCChainRaw struct {
	Contract *QLCChain // Generic contract binding to access the raw methods on
}

// QLCChainCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type QLCChainCallerRaw struct {
	Contract *QLCChainCaller // Generic read-only contract binding to access the raw methods on
}

// QLCChainTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type QLCChainTransactorRaw struct {
	Contract *QLCChainTransactor // Generic write-only contract binding to access the raw methods on
}

// NewQLCChain creates a new instance of QLCChain, bound to a specific deployed contract.
func NewQLCChain(address common.Address, backend bind.ContractBackend) (*QLCChain, error) {
	contract, err := bindQLCChain(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &QLCChain{QLCChainCaller: QLCChainCaller{contract: contract}, QLCChainTransactor: QLCChainTransactor{contract: contract}, QLCChainFilterer: QLCChainFilterer{contract: contract}}, nil
}

// NewQLCChainCaller creates a new read-only instance of QLCChain, bound to a specific deployed contract.
func NewQLCChainCaller(address common.Address, caller bind.ContractCaller) (*QLCChainCaller, error) {
	contract, err := bindQLCChain(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &QLCChainCaller{contract: contract}, nil
}

// NewQLCChainTransactor creates a new write-only instance of QLCChain, bound to a specific deployed contract.
func NewQLCChainTransactor(address common.Address, transactor bind.ContractTransactor) (*QLCChainTransactor, error) {
	contract, err := bindQLCChain(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &QLCChainTransactor{contract: contract}, nil
}

// NewQLCChainFilterer creates a new log filterer instance of QLCChain, bound to a specific deployed contract.
func NewQLCChainFilterer(address common.Address, filterer bind.ContractFilterer) (*QLCChainFilterer, error) {
	contract, err := bindQLCChain(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &QLCChainFilterer{contract: contract}, nil
}

// bindQLCChain binds a generic wrapper to an already deployed contract.
func bindQLCChain(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(QLCChainABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_QLCChain *QLCChainRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _QLCChain.Contract.QLCChainCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_QLCChain *QLCChainRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.Contract.QLCChainTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_QLCChain *QLCChainRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _QLCChain.Contract.QLCChainTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_QLCChain *QLCChainCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _QLCChain.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_QLCChain *QLCChainTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_QLCChain *QLCChainTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _QLCChain.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a paid mutator transaction binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) returns(uint256)
func (_QLCChain *QLCChainTransactor) Allowance(opts *bind.TransactOpts, owner common.Address, spender common.Address) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "allowance", owner, spender)
}

// Allowance is a paid mutator transaction binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) returns(uint256)
func (_QLCChain *QLCChainSession) Allowance(owner common.Address, spender common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.Allowance(&_QLCChain.TransactOpts, owner, spender)
}

// Allowance is a paid mutator transaction binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) returns(uint256)
func (_QLCChain *QLCChainTransactorSession) Allowance(owner common.Address, spender common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.Allowance(&_QLCChain.TransactOpts, owner, spender)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QLCChain *QLCChainSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Approve(&_QLCChain.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Approve(&_QLCChain.TransactOpts, spender, amount)
}

// BalanceOf is a paid mutator transaction binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) returns(uint256)
func (_QLCChain *QLCChainTransactor) BalanceOf(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "balanceOf", account)
}

// BalanceOf is a paid mutator transaction binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) returns(uint256)
func (_QLCChain *QLCChainSession) BalanceOf(account common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.BalanceOf(&_QLCChain.TransactOpts, account)
}

// BalanceOf is a paid mutator transaction binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) returns(uint256)
func (_QLCChain *QLCChainTransactorSession) BalanceOf(account common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.BalanceOf(&_QLCChain.TransactOpts, account)
}

// Decimals is a paid mutator transaction binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_QLCChain *QLCChainTransactor) Decimals(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "decimals")
}

// Decimals is a paid mutator transaction binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_QLCChain *QLCChainSession) Decimals() (*types.Transaction, error) {
	return _QLCChain.Contract.Decimals(&_QLCChain.TransactOpts)
}

// Decimals is a paid mutator transaction binding the contract method 0x313ce567.
//
// Solidity: function decimals() returns(uint8)
func (_QLCChain *QLCChainTransactorSession) Decimals() (*types.Transaction, error) {
	return _QLCChain.Contract.Decimals(&_QLCChain.TransactOpts)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QLCChain *QLCChainTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QLCChain *QLCChainSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.DecreaseAllowance(&_QLCChain.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QLCChain *QLCChainTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.DecreaseAllowance(&_QLCChain.TransactOpts, spender, subtractedValue)
}

// DestoryFetch is a paid mutator transaction binding the contract method 0x3990ebff.
//
// Solidity: function destoryFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainTransactor) DestoryFetch(opts *bind.TransactOpts, rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "destoryFetch", rHash)
}

// DestoryFetch is a paid mutator transaction binding the contract method 0x3990ebff.
//
// Solidity: function destoryFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainSession) DestoryFetch(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryFetch(&_QLCChain.TransactOpts, rHash)
}

// DestoryFetch is a paid mutator transaction binding the contract method 0x3990ebff.
//
// Solidity: function destoryFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainTransactorSession) DestoryFetch(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryFetch(&_QLCChain.TransactOpts, rHash)
}

// DestoryLock is a paid mutator transaction binding the contract method 0xd067c425.
//
// Solidity: function destoryLock(bytes32 rHash, uint256 amount, address executor) returns()
func (_QLCChain *QLCChainTransactor) DestoryLock(opts *bind.TransactOpts, rHash [32]byte, amount *big.Int, executor common.Address) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "destoryLock", rHash, amount, executor)
}

// DestoryLock is a paid mutator transaction binding the contract method 0xd067c425.
//
// Solidity: function destoryLock(bytes32 rHash, uint256 amount, address executor) returns()
func (_QLCChain *QLCChainSession) DestoryLock(rHash [32]byte, amount *big.Int, executor common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryLock(&_QLCChain.TransactOpts, rHash, amount, executor)
}

// DestoryLock is a paid mutator transaction binding the contract method 0xd067c425.
//
// Solidity: function destoryLock(bytes32 rHash, uint256 amount, address executor) returns()
func (_QLCChain *QLCChainTransactorSession) DestoryLock(rHash [32]byte, amount *big.Int, executor common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryLock(&_QLCChain.TransactOpts, rHash, amount, executor)
}

// DestoryUnlock is a paid mutator transaction binding the contract method 0x25984de3.
//
// Solidity: function destoryUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainTransactor) DestoryUnlock(opts *bind.TransactOpts, rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "destoryUnlock", rHash, rOrigin)
}

// DestoryUnlock is a paid mutator transaction binding the contract method 0x25984de3.
//
// Solidity: function destoryUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainSession) DestoryUnlock(rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryUnlock(&_QLCChain.TransactOpts, rHash, rOrigin)
}

// DestoryUnlock is a paid mutator transaction binding the contract method 0x25984de3.
//
// Solidity: function destoryUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainTransactorSession) DestoryUnlock(rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.DestoryUnlock(&_QLCChain.TransactOpts, rHash, rOrigin)
}

// HashTimer is a paid mutator transaction binding the contract method 0x6aacd506.
//
// Solidity: function hashTimer(bytes32 rHash) returns(bytes32, uint256, address, uint256, uint256)
func (_QLCChain *QLCChainTransactor) HashTimer(opts *bind.TransactOpts, rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "hashTimer", rHash)
}

// HashTimer is a paid mutator transaction binding the contract method 0x6aacd506.
//
// Solidity: function hashTimer(bytes32 rHash) returns(bytes32, uint256, address, uint256, uint256)
func (_QLCChain *QLCChainSession) HashTimer(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.HashTimer(&_QLCChain.TransactOpts, rHash)
}

// HashTimer is a paid mutator transaction binding the contract method 0x6aacd506.
//
// Solidity: function hashTimer(bytes32 rHash) returns(bytes32, uint256, address, uint256, uint256)
func (_QLCChain *QLCChainTransactorSession) HashTimer(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.HashTimer(&_QLCChain.TransactOpts, rHash)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QLCChain *QLCChainTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QLCChain *QLCChainSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.IncreaseAllowance(&_QLCChain.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QLCChain *QLCChainTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.IncreaseAllowance(&_QLCChain.TransactOpts, spender, addedValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QLCChain *QLCChainTransactor) Initialize(opts *bind.TransactOpts, name string, symbol string) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "initialize", name, symbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QLCChain *QLCChainSession) Initialize(name string, symbol string) (*types.Transaction, error) {
	return _QLCChain.Contract.Initialize(&_QLCChain.TransactOpts, name, symbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QLCChain *QLCChainTransactorSession) Initialize(name string, symbol string) (*types.Transaction, error) {
	return _QLCChain.Contract.Initialize(&_QLCChain.TransactOpts, name, symbol)
}

// IssueFetch is a paid mutator transaction binding the contract method 0x19a4440a.
//
// Solidity: function issueFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainTransactor) IssueFetch(opts *bind.TransactOpts, rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "issueFetch", rHash)
}

// IssueFetch is a paid mutator transaction binding the contract method 0x19a4440a.
//
// Solidity: function issueFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainSession) IssueFetch(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueFetch(&_QLCChain.TransactOpts, rHash)
}

// IssueFetch is a paid mutator transaction binding the contract method 0x19a4440a.
//
// Solidity: function issueFetch(bytes32 rHash) returns()
func (_QLCChain *QLCChainTransactorSession) IssueFetch(rHash [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueFetch(&_QLCChain.TransactOpts, rHash)
}

// IssueLock is a paid mutator transaction binding the contract method 0xdd049cd0.
//
// Solidity: function issueLock(bytes32 rHash, uint256 amount) returns()
func (_QLCChain *QLCChainTransactor) IssueLock(opts *bind.TransactOpts, rHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "issueLock", rHash, amount)
}

// IssueLock is a paid mutator transaction binding the contract method 0xdd049cd0.
//
// Solidity: function issueLock(bytes32 rHash, uint256 amount) returns()
func (_QLCChain *QLCChainSession) IssueLock(rHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueLock(&_QLCChain.TransactOpts, rHash, amount)
}

// IssueLock is a paid mutator transaction binding the contract method 0xdd049cd0.
//
// Solidity: function issueLock(bytes32 rHash, uint256 amount) returns()
func (_QLCChain *QLCChainTransactorSession) IssueLock(rHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueLock(&_QLCChain.TransactOpts, rHash, amount)
}

// IssueUnlock is a paid mutator transaction binding the contract method 0x501f18f8.
//
// Solidity: function issueUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainTransactor) IssueUnlock(opts *bind.TransactOpts, rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "issueUnlock", rHash, rOrigin)
}

// IssueUnlock is a paid mutator transaction binding the contract method 0x501f18f8.
//
// Solidity: function issueUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainSession) IssueUnlock(rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueUnlock(&_QLCChain.TransactOpts, rHash, rOrigin)
}

// IssueUnlock is a paid mutator transaction binding the contract method 0x501f18f8.
//
// Solidity: function issueUnlock(bytes32 rHash, bytes32 rOrigin) returns()
func (_QLCChain *QLCChainTransactorSession) IssueUnlock(rHash [32]byte, rOrigin [32]byte) (*types.Transaction, error) {
	return _QLCChain.Contract.IssueUnlock(&_QLCChain.TransactOpts, rHash, rOrigin)
}

// Name is a paid mutator transaction binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_QLCChain *QLCChainTransactor) Name(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "name")
}

// Name is a paid mutator transaction binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_QLCChain *QLCChainSession) Name() (*types.Transaction, error) {
	return _QLCChain.Contract.Name(&_QLCChain.TransactOpts)
}

// Name is a paid mutator transaction binding the contract method 0x06fdde03.
//
// Solidity: function name() returns(string)
func (_QLCChain *QLCChainTransactorSession) Name() (*types.Transaction, error) {
	return _QLCChain.Contract.Name(&_QLCChain.TransactOpts)
}

// Owner is a paid mutator transaction binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_QLCChain *QLCChainTransactor) Owner(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "owner")
}

// Owner is a paid mutator transaction binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_QLCChain *QLCChainSession) Owner() (*types.Transaction, error) {
	return _QLCChain.Contract.Owner(&_QLCChain.TransactOpts)
}

// Owner is a paid mutator transaction binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_QLCChain *QLCChainTransactorSession) Owner() (*types.Transaction, error) {
	return _QLCChain.Contract.Owner(&_QLCChain.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QLCChain *QLCChainTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QLCChain *QLCChainSession) RenounceOwnership() (*types.Transaction, error) {
	return _QLCChain.Contract.RenounceOwnership(&_QLCChain.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QLCChain *QLCChainTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _QLCChain.Contract.RenounceOwnership(&_QLCChain.TransactOpts)
}

// Symbol is a paid mutator transaction binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_QLCChain *QLCChainTransactor) Symbol(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "symbol")
}

// Symbol is a paid mutator transaction binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_QLCChain *QLCChainSession) Symbol() (*types.Transaction, error) {
	return _QLCChain.Contract.Symbol(&_QLCChain.TransactOpts)
}

// Symbol is a paid mutator transaction binding the contract method 0x95d89b41.
//
// Solidity: function symbol() returns(string)
func (_QLCChain *QLCChainTransactorSession) Symbol() (*types.Transaction, error) {
	return _QLCChain.Contract.Symbol(&_QLCChain.TransactOpts)
}

// TotalSupply is a paid mutator transaction binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_QLCChain *QLCChainTransactor) TotalSupply(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "totalSupply")
}

// TotalSupply is a paid mutator transaction binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_QLCChain *QLCChainSession) TotalSupply() (*types.Transaction, error) {
	return _QLCChain.Contract.TotalSupply(&_QLCChain.TransactOpts)
}

// TotalSupply is a paid mutator transaction binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() returns(uint256)
func (_QLCChain *QLCChainTransactorSession) TotalSupply() (*types.Transaction, error) {
	return _QLCChain.Contract.TotalSupply(&_QLCChain.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Transfer(&_QLCChain.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Transfer(&_QLCChain.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.TransferFrom(&_QLCChain.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QLCChain *QLCChainTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.TransferFrom(&_QLCChain.TransactOpts, sender, recipient, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QLCChain *QLCChainTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QLCChain *QLCChainSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.TransferOwnership(&_QLCChain.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QLCChain *QLCChainTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _QLCChain.Contract.TransferOwnership(&_QLCChain.TransactOpts, newOwner)
}

// QLCChainApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the QLCChain contract.
type QLCChainApprovalIterator struct {
	Event *QLCChainApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *QLCChainApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(QLCChainApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *QLCChainApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainApproval represents a Approval event raised by the QLCChain contract.
type QLCChainApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_QLCChain *QLCChainFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*QLCChainApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainApprovalIterator{contract: _QLCChain.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_QLCChain *QLCChainFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *QLCChainApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainApproval)
				if err := _QLCChain.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_QLCChain *QLCChainFilterer) ParseApproval(log types.Log) (*QLCChainApproval, error) {
	event := new(QLCChainApproval)
	if err := _QLCChain.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	return event, nil
}

// QLCChainLockedStateIterator is returned from FilterLockedState and is used to iterate over the raw logs and unpacked data for LockedState events raised by the QLCChain contract.
type QLCChainLockedStateIterator struct {
	Event *QLCChainLockedState // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *QLCChainLockedStateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainLockedState)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(QLCChainLockedState)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *QLCChainLockedStateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainLockedStateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainLockedState represents a LockedState event raised by the QLCChain contract.
type QLCChainLockedState struct {
	RHash [32]byte
	State *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterLockedState is a free log retrieval operation binding the contract event 0x9602218484dbca102b1b8ecd40a2b8d3a19f098859a193580428927b239737db.
//
// Solidity: event LockedState(bytes32 indexed rHash, uint256 state)
func (_QLCChain *QLCChainFilterer) FilterLockedState(opts *bind.FilterOpts, rHash [][32]byte) (*QLCChainLockedStateIterator, error) {

	var rHashRule []interface{}
	for _, rHashItem := range rHash {
		rHashRule = append(rHashRule, rHashItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "LockedState", rHashRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainLockedStateIterator{contract: _QLCChain.contract, event: "LockedState", logs: logs, sub: sub}, nil
}

// WatchLockedState is a free log subscription operation binding the contract event 0x9602218484dbca102b1b8ecd40a2b8d3a19f098859a193580428927b239737db.
//
// Solidity: event LockedState(bytes32 indexed rHash, uint256 state)
func (_QLCChain *QLCChainFilterer) WatchLockedState(opts *bind.WatchOpts, sink chan<- *QLCChainLockedState, rHash [][32]byte) (event.Subscription, error) {

	var rHashRule []interface{}
	for _, rHashItem := range rHash {
		rHashRule = append(rHashRule, rHashItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "LockedState", rHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainLockedState)
				if err := _QLCChain.contract.UnpackLog(event, "LockedState", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLockedState is a log parse operation binding the contract event 0x9602218484dbca102b1b8ecd40a2b8d3a19f098859a193580428927b239737db.
//
// Solidity: event LockedState(bytes32 indexed rHash, uint256 state)
func (_QLCChain *QLCChainFilterer) ParseLockedState(log types.Log) (*QLCChainLockedState, error) {
	event := new(QLCChainLockedState)
	if err := _QLCChain.contract.UnpackLog(event, "LockedState", log); err != nil {
		return nil, err
	}
	return event, nil
}

// QLCChainOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the QLCChain contract.
type QLCChainOwnershipTransferredIterator struct {
	Event *QLCChainOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *QLCChainOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(QLCChainOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *QLCChainOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainOwnershipTransferred represents a OwnershipTransferred event raised by the QLCChain contract.
type QLCChainOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_QLCChain *QLCChainFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*QLCChainOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainOwnershipTransferredIterator{contract: _QLCChain.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_QLCChain *QLCChainFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *QLCChainOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainOwnershipTransferred)
				if err := _QLCChain.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_QLCChain *QLCChainFilterer) ParseOwnershipTransferred(log types.Log) (*QLCChainOwnershipTransferred, error) {
	event := new(QLCChainOwnershipTransferred)
	if err := _QLCChain.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	return event, nil
}

// QLCChainTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the QLCChain contract.
type QLCChainTransferIterator struct {
	Event *QLCChainTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *QLCChainTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(QLCChainTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *QLCChainTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainTransfer represents a Transfer event raised by the QLCChain contract.
type QLCChainTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_QLCChain *QLCChainFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*QLCChainTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainTransferIterator{contract: _QLCChain.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_QLCChain *QLCChainFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *QLCChainTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainTransfer)
				if err := _QLCChain.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_QLCChain *QLCChainFilterer) ParseTransfer(log types.Log) (*QLCChainTransfer, error) {
	event := new(QLCChainTransfer)
	if err := _QLCChain.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	return event, nil
}
