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
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// QLCChainABI is the input ABI used to generate the binding from.
const QLCChainABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nep5Addr\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"nep5Hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"active\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"lockedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"nep5Hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"nep5Addr\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"circuitBraker\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

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
func (_QLCChain *QLCChainRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
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
func (_QLCChain *QLCChainCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
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

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QLCChain *QLCChainCaller) Active(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "active")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QLCChain *QLCChainSession) Active() (bool, error) {
	return _QLCChain.Contract.Active(&_QLCChain.CallOpts)
}

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QLCChain *QLCChainCallerSession) Active() (bool, error) {
	return _QLCChain.Contract.Active(&_QLCChain.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QLCChain *QLCChainCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QLCChain *QLCChainSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _QLCChain.Contract.Allowance(&_QLCChain.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QLCChain *QLCChainCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _QLCChain.Contract.Allowance(&_QLCChain.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QLCChain *QLCChainCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QLCChain *QLCChainSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _QLCChain.Contract.BalanceOf(&_QLCChain.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QLCChain *QLCChainCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _QLCChain.Contract.BalanceOf(&_QLCChain.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QLCChain *QLCChainCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QLCChain *QLCChainSession) Decimals() (uint8, error) {
	return _QLCChain.Contract.Decimals(&_QLCChain.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QLCChain *QLCChainCallerSession) Decimals() (uint8, error) {
	return _QLCChain.Contract.Decimals(&_QLCChain.CallOpts)
}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QLCChain *QLCChainCaller) LockedAmount(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "lockedAmount", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QLCChain *QLCChainSession) LockedAmount(arg0 [32]byte) (*big.Int, error) {
	return _QLCChain.Contract.LockedAmount(&_QLCChain.CallOpts, arg0)
}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QLCChain *QLCChainCallerSession) LockedAmount(arg0 [32]byte) (*big.Int, error) {
	return _QLCChain.Contract.LockedAmount(&_QLCChain.CallOpts, arg0)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QLCChain *QLCChainCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QLCChain *QLCChainSession) Name() (string, error) {
	return _QLCChain.Contract.Name(&_QLCChain.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QLCChain *QLCChainCallerSession) Name() (string, error) {
	return _QLCChain.Contract.Name(&_QLCChain.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QLCChain *QLCChainCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QLCChain *QLCChainSession) Owner() (common.Address, error) {
	return _QLCChain.Contract.Owner(&_QLCChain.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QLCChain *QLCChainCallerSession) Owner() (common.Address, error) {
	return _QLCChain.Contract.Owner(&_QLCChain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QLCChain *QLCChainCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QLCChain *QLCChainSession) Symbol() (string, error) {
	return _QLCChain.Contract.Symbol(&_QLCChain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QLCChain *QLCChainCallerSession) Symbol() (string, error) {
	return _QLCChain.Contract.Symbol(&_QLCChain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QLCChain *QLCChainCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _QLCChain.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QLCChain *QLCChainSession) TotalSupply() (*big.Int, error) {
	return _QLCChain.Contract.TotalSupply(&_QLCChain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QLCChain *QLCChainCallerSession) TotalSupply() (*big.Int, error) {
	return _QLCChain.Contract.TotalSupply(&_QLCChain.CallOpts)
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

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string nep5Addr, uint256 amount) returns()
func (_QLCChain *QLCChainTransactor) Burn(opts *bind.TransactOpts, nep5Addr string, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "burn", nep5Addr, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string nep5Addr, uint256 amount) returns()
func (_QLCChain *QLCChainSession) Burn(nep5Addr string, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Burn(&_QLCChain.TransactOpts, nep5Addr, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string nep5Addr, uint256 amount) returns()
func (_QLCChain *QLCChainTransactorSession) Burn(nep5Addr string, amount *big.Int) (*types.Transaction, error) {
	return _QLCChain.Contract.Burn(&_QLCChain.TransactOpts, nep5Addr, amount)
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QLCChain *QLCChainTransactor) CircuitBraker(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "circuitBraker")
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QLCChain *QLCChainSession) CircuitBraker() (*types.Transaction, error) {
	return _QLCChain.Contract.CircuitBraker(&_QLCChain.TransactOpts)
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QLCChain *QLCChainTransactorSession) CircuitBraker() (*types.Transaction, error) {
	return _QLCChain.Contract.CircuitBraker(&_QLCChain.TransactOpts)
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

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 nep5Hash, bytes signature) returns()
func (_QLCChain *QLCChainTransactor) Mint(opts *bind.TransactOpts, amount *big.Int, nep5Hash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QLCChain.contract.Transact(opts, "mint", amount, nep5Hash, signature)
}

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 nep5Hash, bytes signature) returns()
func (_QLCChain *QLCChainSession) Mint(amount *big.Int, nep5Hash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QLCChain.Contract.Mint(&_QLCChain.TransactOpts, amount, nep5Hash, signature)
}

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 nep5Hash, bytes signature) returns()
func (_QLCChain *QLCChainTransactorSession) Mint(amount *big.Int, nep5Hash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QLCChain.Contract.Mint(&_QLCChain.TransactOpts, amount, nep5Hash, signature)
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
	event.Raw = log
	return event, nil
}

// QLCChainBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the QLCChain contract.
type QLCChainBurnIterator struct {
	Event *QLCChainBurn // Event containing the contract specifics and raw log

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
func (it *QLCChainBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainBurn)
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
		it.Event = new(QLCChainBurn)
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
func (it *QLCChainBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainBurn represents a Burn event raised by the QLCChain contract.
type QLCChainBurn struct {
	User     common.Address
	Nep5Addr string
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xfdf096248d2b7b0aef506231c043107c21faacc26193881b3f0cdc8b5479692a.
//
// Solidity: event Burn(address indexed user, string nep5Addr, uint256 amount)
func (_QLCChain *QLCChainFilterer) FilterBurn(opts *bind.FilterOpts, user []common.Address) (*QLCChainBurnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "Burn", userRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainBurnIterator{contract: _QLCChain.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xfdf096248d2b7b0aef506231c043107c21faacc26193881b3f0cdc8b5479692a.
//
// Solidity: event Burn(address indexed user, string nep5Addr, uint256 amount)
func (_QLCChain *QLCChainFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *QLCChainBurn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "Burn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainBurn)
				if err := _QLCChain.contract.UnpackLog(event, "Burn", log); err != nil {
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

// ParseBurn is a log parse operation binding the contract event 0xfdf096248d2b7b0aef506231c043107c21faacc26193881b3f0cdc8b5479692a.
//
// Solidity: event Burn(address indexed user, string nep5Addr, uint256 amount)
func (_QLCChain *QLCChainFilterer) ParseBurn(log types.Log) (*QLCChainBurn, error) {
	event := new(QLCChainBurn)
	if err := _QLCChain.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// QLCChainMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the QLCChain contract.
type QLCChainMintIterator struct {
	Event *QLCChainMint // Event containing the contract specifics and raw log

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
func (it *QLCChainMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QLCChainMint)
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
		it.Event = new(QLCChainMint)
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
func (it *QLCChainMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QLCChainMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QLCChainMint represents a Mint event raised by the QLCChain contract.
type QLCChainMint struct {
	User     common.Address
	Nep5Hash [32]byte
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x103a2d32aec953695f3b9ec5ed6c1c6cb822debe92cf1fcf0832cb2c262c7eec.
//
// Solidity: event Mint(address indexed user, bytes32 nep5Hash, uint256 amount)
func (_QLCChain *QLCChainFilterer) FilterMint(opts *bind.FilterOpts, user []common.Address) (*QLCChainMintIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QLCChain.contract.FilterLogs(opts, "Mint", userRule)
	if err != nil {
		return nil, err
	}
	return &QLCChainMintIterator{contract: _QLCChain.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x103a2d32aec953695f3b9ec5ed6c1c6cb822debe92cf1fcf0832cb2c262c7eec.
//
// Solidity: event Mint(address indexed user, bytes32 nep5Hash, uint256 amount)
func (_QLCChain *QLCChainFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *QLCChainMint, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QLCChain.contract.WatchLogs(opts, "Mint", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QLCChainMint)
				if err := _QLCChain.contract.UnpackLog(event, "Mint", log); err != nil {
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

// ParseMint is a log parse operation binding the contract event 0x103a2d32aec953695f3b9ec5ed6c1c6cb822debe92cf1fcf0832cb2c262c7eec.
//
// Solidity: event Mint(address indexed user, bytes32 nep5Hash, uint256 amount)
func (_QLCChain *QLCChainFilterer) ParseMint(log types.Log) (*QLCChainMint, error) {
	event := new(QLCChainMint)
	if err := _QLCChain.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
	return event, nil
}
