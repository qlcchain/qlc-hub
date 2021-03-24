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

// QGasChainABI is the input ABI used to generate the binding from.
const QGasChainABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"qlcAddr\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"qlcHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"active\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"lockedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"qlcHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"qlcAddr\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"circuitBraker\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// QGasChain is an auto generated Go binding around an Ethereum contract.
type QGasChain struct {
	QGasChainCaller     // Read-only binding to the contract
	QGasChainTransactor // Write-only binding to the contract
	QGasChainFilterer   // Log filterer for contract events
}

// QGasChainCaller is an auto generated read-only Go binding around an Ethereum contract.
type QGasChainCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QGasChainTransactor is an auto generated write-only Go binding around an Ethereum contract.
type QGasChainTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QGasChainFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type QGasChainFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QGasChainSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type QGasChainSession struct {
	Contract     *QGasChain        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// QGasChainCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type QGasChainCallerSession struct {
	Contract *QGasChainCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// QGasChainTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type QGasChainTransactorSession struct {
	Contract     *QGasChainTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// QGasChainRaw is an auto generated low-level Go binding around an Ethereum contract.
type QGasChainRaw struct {
	Contract *QGasChain // Generic contract binding to access the raw methods on
}

// QGasChainCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type QGasChainCallerRaw struct {
	Contract *QGasChainCaller // Generic read-only contract binding to access the raw methods on
}

// QGasChainTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type QGasChainTransactorRaw struct {
	Contract *QGasChainTransactor // Generic write-only contract binding to access the raw methods on
}

// NewQGasChain creates a new instance of QGasChain, bound to a specific deployed contract.
func NewQGasChain(address common.Address, backend bind.ContractBackend) (*QGasChain, error) {
	contract, err := bindQGasChain(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &QGasChain{QGasChainCaller: QGasChainCaller{contract: contract}, QGasChainTransactor: QGasChainTransactor{contract: contract}, QGasChainFilterer: QGasChainFilterer{contract: contract}}, nil
}

// NewQGasChainCaller creates a new read-only instance of QGasChain, bound to a specific deployed contract.
func NewQGasChainCaller(address common.Address, caller bind.ContractCaller) (*QGasChainCaller, error) {
	contract, err := bindQGasChain(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &QGasChainCaller{contract: contract}, nil
}

// NewQGasChainTransactor creates a new write-only instance of QGasChain, bound to a specific deployed contract.
func NewQGasChainTransactor(address common.Address, transactor bind.ContractTransactor) (*QGasChainTransactor, error) {
	contract, err := bindQGasChain(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &QGasChainTransactor{contract: contract}, nil
}

// NewQGasChainFilterer creates a new log filterer instance of QGasChain, bound to a specific deployed contract.
func NewQGasChainFilterer(address common.Address, filterer bind.ContractFilterer) (*QGasChainFilterer, error) {
	contract, err := bindQGasChain(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &QGasChainFilterer{contract: contract}, nil
}

// bindQGasChain binds a generic wrapper to an already deployed contract.
func bindQGasChain(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(QGasChainABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_QGasChain *QGasChainRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _QGasChain.Contract.QGasChainCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_QGasChain *QGasChainRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QGasChain.Contract.QGasChainTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_QGasChain *QGasChainRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _QGasChain.Contract.QGasChainTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_QGasChain *QGasChainCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _QGasChain.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_QGasChain *QGasChainTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QGasChain.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_QGasChain *QGasChainTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _QGasChain.Contract.contract.Transact(opts, method, params...)
}

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QGasChain *QGasChainCaller) Active(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "active")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QGasChain *QGasChainSession) Active() (bool, error) {
	return _QGasChain.Contract.Active(&_QGasChain.CallOpts)
}

// Active is a free data retrieval call binding the contract method 0x02fb0c5e.
//
// Solidity: function active() view returns(bool)
func (_QGasChain *QGasChainCallerSession) Active() (bool, error) {
	return _QGasChain.Contract.Active(&_QGasChain.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QGasChain *QGasChainCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QGasChain *QGasChainSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _QGasChain.Contract.Allowance(&_QGasChain.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_QGasChain *QGasChainCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _QGasChain.Contract.Allowance(&_QGasChain.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QGasChain *QGasChainCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QGasChain *QGasChainSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _QGasChain.Contract.BalanceOf(&_QGasChain.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_QGasChain *QGasChainCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _QGasChain.Contract.BalanceOf(&_QGasChain.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QGasChain *QGasChainCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QGasChain *QGasChainSession) Decimals() (uint8, error) {
	return _QGasChain.Contract.Decimals(&_QGasChain.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_QGasChain *QGasChainCallerSession) Decimals() (uint8, error) {
	return _QGasChain.Contract.Decimals(&_QGasChain.CallOpts)
}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QGasChain *QGasChainCaller) LockedAmount(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "lockedAmount", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QGasChain *QGasChainSession) LockedAmount(arg0 [32]byte) (*big.Int, error) {
	return _QGasChain.Contract.LockedAmount(&_QGasChain.CallOpts, arg0)
}

// LockedAmount is a free data retrieval call binding the contract method 0x172a16a4.
//
// Solidity: function lockedAmount(bytes32 ) view returns(uint256)
func (_QGasChain *QGasChainCallerSession) LockedAmount(arg0 [32]byte) (*big.Int, error) {
	return _QGasChain.Contract.LockedAmount(&_QGasChain.CallOpts, arg0)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QGasChain *QGasChainCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QGasChain *QGasChainSession) Name() (string, error) {
	return _QGasChain.Contract.Name(&_QGasChain.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_QGasChain *QGasChainCallerSession) Name() (string, error) {
	return _QGasChain.Contract.Name(&_QGasChain.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QGasChain *QGasChainCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QGasChain *QGasChainSession) Owner() (common.Address, error) {
	return _QGasChain.Contract.Owner(&_QGasChain.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_QGasChain *QGasChainCallerSession) Owner() (common.Address, error) {
	return _QGasChain.Contract.Owner(&_QGasChain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QGasChain *QGasChainCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QGasChain *QGasChainSession) Symbol() (string, error) {
	return _QGasChain.Contract.Symbol(&_QGasChain.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_QGasChain *QGasChainCallerSession) Symbol() (string, error) {
	return _QGasChain.Contract.Symbol(&_QGasChain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QGasChain *QGasChainCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _QGasChain.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QGasChain *QGasChainSession) TotalSupply() (*big.Int, error) {
	return _QGasChain.Contract.TotalSupply(&_QGasChain.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_QGasChain *QGasChainCallerSession) TotalSupply() (*big.Int, error) {
	return _QGasChain.Contract.TotalSupply(&_QGasChain.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QGasChain *QGasChainSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Approve(&_QGasChain.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Approve(&_QGasChain.TransactOpts, spender, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string qlcAddr, uint256 amount) returns()
func (_QGasChain *QGasChainTransactor) Burn(opts *bind.TransactOpts, qlcAddr string, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "burn", qlcAddr, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string qlcAddr, uint256 amount) returns()
func (_QGasChain *QGasChainSession) Burn(qlcAddr string, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Burn(&_QGasChain.TransactOpts, qlcAddr, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb48272cc.
//
// Solidity: function burn(string qlcAddr, uint256 amount) returns()
func (_QGasChain *QGasChainTransactorSession) Burn(qlcAddr string, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Burn(&_QGasChain.TransactOpts, qlcAddr, amount)
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QGasChain *QGasChainTransactor) CircuitBraker(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "circuitBraker")
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QGasChain *QGasChainSession) CircuitBraker() (*types.Transaction, error) {
	return _QGasChain.Contract.CircuitBraker(&_QGasChain.TransactOpts)
}

// CircuitBraker is a paid mutator transaction binding the contract method 0xdd064a7d.
//
// Solidity: function circuitBraker() returns()
func (_QGasChain *QGasChainTransactorSession) CircuitBraker() (*types.Transaction, error) {
	return _QGasChain.Contract.CircuitBraker(&_QGasChain.TransactOpts)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QGasChain *QGasChainTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QGasChain *QGasChainSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.DecreaseAllowance(&_QGasChain.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_QGasChain *QGasChainTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.DecreaseAllowance(&_QGasChain.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QGasChain *QGasChainTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QGasChain *QGasChainSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.IncreaseAllowance(&_QGasChain.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_QGasChain *QGasChainTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.IncreaseAllowance(&_QGasChain.TransactOpts, spender, addedValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QGasChain *QGasChainTransactor) Initialize(opts *bind.TransactOpts, name string, symbol string) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "initialize", name, symbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QGasChain *QGasChainSession) Initialize(name string, symbol string) (*types.Transaction, error) {
	return _QGasChain.Contract.Initialize(&_QGasChain.TransactOpts, name, symbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string symbol) returns()
func (_QGasChain *QGasChainTransactorSession) Initialize(name string, symbol string) (*types.Transaction, error) {
	return _QGasChain.Contract.Initialize(&_QGasChain.TransactOpts, name, symbol)
}

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 qlcHash, bytes signature) returns()
func (_QGasChain *QGasChainTransactor) Mint(opts *bind.TransactOpts, amount *big.Int, qlcHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "mint", amount, qlcHash, signature)
}

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 qlcHash, bytes signature) returns()
func (_QGasChain *QGasChainSession) Mint(amount *big.Int, qlcHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QGasChain.Contract.Mint(&_QGasChain.TransactOpts, amount, qlcHash, signature)
}

// Mint is a paid mutator transaction binding the contract method 0x9ab475b5.
//
// Solidity: function mint(uint256 amount, bytes32 qlcHash, bytes signature) returns()
func (_QGasChain *QGasChainTransactorSession) Mint(amount *big.Int, qlcHash [32]byte, signature []byte) (*types.Transaction, error) {
	return _QGasChain.Contract.Mint(&_QGasChain.TransactOpts, amount, qlcHash, signature)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QGasChain *QGasChainTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QGasChain *QGasChainSession) RenounceOwnership() (*types.Transaction, error) {
	return _QGasChain.Contract.RenounceOwnership(&_QGasChain.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_QGasChain *QGasChainTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _QGasChain.Contract.RenounceOwnership(&_QGasChain.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Transfer(&_QGasChain.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.Transfer(&_QGasChain.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.TransferFrom(&_QGasChain.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_QGasChain *QGasChainTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _QGasChain.Contract.TransferFrom(&_QGasChain.TransactOpts, sender, recipient, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QGasChain *QGasChainTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _QGasChain.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QGasChain *QGasChainSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _QGasChain.Contract.TransferOwnership(&_QGasChain.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_QGasChain *QGasChainTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _QGasChain.Contract.TransferOwnership(&_QGasChain.TransactOpts, newOwner)
}

// QGasChainApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the QGasChain contract.
type QGasChainApprovalIterator struct {
	Event *QGasChainApproval // Event containing the contract specifics and raw log

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
func (it *QGasChainApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QGasChainApproval)
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
		it.Event = new(QGasChainApproval)
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
func (it *QGasChainApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QGasChainApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QGasChainApproval represents a Approval event raised by the QGasChain contract.
type QGasChainApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_QGasChain *QGasChainFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*QGasChainApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _QGasChain.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &QGasChainApprovalIterator{contract: _QGasChain.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_QGasChain *QGasChainFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *QGasChainApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _QGasChain.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QGasChainApproval)
				if err := _QGasChain.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_QGasChain *QGasChainFilterer) ParseApproval(log types.Log) (*QGasChainApproval, error) {
	event := new(QGasChainApproval)
	if err := _QGasChain.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// QGasChainBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the QGasChain contract.
type QGasChainBurnIterator struct {
	Event *QGasChainBurn // Event containing the contract specifics and raw log

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
func (it *QGasChainBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QGasChainBurn)
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
		it.Event = new(QGasChainBurn)
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
func (it *QGasChainBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QGasChainBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QGasChainBurn represents a Burn event raised by the QGasChain contract.
type QGasChainBurn struct {
	User    common.Address
	QlcAddr string
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xfdf096248d2b7b0aef506231c043107c21faacc26193881b3f0cdc8b5479692a.
//
// Solidity: event Burn(address indexed user, string qlcAddr, uint256 amount)
func (_QGasChain *QGasChainFilterer) FilterBurn(opts *bind.FilterOpts, user []common.Address) (*QGasChainBurnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QGasChain.contract.FilterLogs(opts, "Burn", userRule)
	if err != nil {
		return nil, err
	}
	return &QGasChainBurnIterator{contract: _QGasChain.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xfdf096248d2b7b0aef506231c043107c21faacc26193881b3f0cdc8b5479692a.
//
// Solidity: event Burn(address indexed user, string qlcAddr, uint256 amount)
func (_QGasChain *QGasChainFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *QGasChainBurn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QGasChain.contract.WatchLogs(opts, "Burn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QGasChainBurn)
				if err := _QGasChain.contract.UnpackLog(event, "Burn", log); err != nil {
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
// Solidity: event Burn(address indexed user, string qlcAddr, uint256 amount)
func (_QGasChain *QGasChainFilterer) ParseBurn(log types.Log) (*QGasChainBurn, error) {
	event := new(QGasChainBurn)
	if err := _QGasChain.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// QGasChainMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the QGasChain contract.
type QGasChainMintIterator struct {
	Event *QGasChainMint // Event containing the contract specifics and raw log

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
func (it *QGasChainMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QGasChainMint)
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
		it.Event = new(QGasChainMint)
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
func (it *QGasChainMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QGasChainMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QGasChainMint represents a Mint event raised by the QGasChain contract.
type QGasChainMint struct {
	User    common.Address
	QlcHash [32]byte
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x103a2d32aec953695f3b9ec5ed6c1c6cb822debe92cf1fcf0832cb2c262c7eec.
//
// Solidity: event Mint(address indexed user, bytes32 qlcHash, uint256 amount)
func (_QGasChain *QGasChainFilterer) FilterMint(opts *bind.FilterOpts, user []common.Address) (*QGasChainMintIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QGasChain.contract.FilterLogs(opts, "Mint", userRule)
	if err != nil {
		return nil, err
	}
	return &QGasChainMintIterator{contract: _QGasChain.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x103a2d32aec953695f3b9ec5ed6c1c6cb822debe92cf1fcf0832cb2c262c7eec.
//
// Solidity: event Mint(address indexed user, bytes32 qlcHash, uint256 amount)
func (_QGasChain *QGasChainFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *QGasChainMint, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _QGasChain.contract.WatchLogs(opts, "Mint", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QGasChainMint)
				if err := _QGasChain.contract.UnpackLog(event, "Mint", log); err != nil {
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
// Solidity: event Mint(address indexed user, bytes32 qlcHash, uint256 amount)
func (_QGasChain *QGasChainFilterer) ParseMint(log types.Log) (*QGasChainMint, error) {
	event := new(QGasChainMint)
	if err := _QGasChain.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// QGasChainOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the QGasChain contract.
type QGasChainOwnershipTransferredIterator struct {
	Event *QGasChainOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *QGasChainOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QGasChainOwnershipTransferred)
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
		it.Event = new(QGasChainOwnershipTransferred)
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
func (it *QGasChainOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QGasChainOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QGasChainOwnershipTransferred represents a OwnershipTransferred event raised by the QGasChain contract.
type QGasChainOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_QGasChain *QGasChainFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*QGasChainOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _QGasChain.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &QGasChainOwnershipTransferredIterator{contract: _QGasChain.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_QGasChain *QGasChainFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *QGasChainOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _QGasChain.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QGasChainOwnershipTransferred)
				if err := _QGasChain.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_QGasChain *QGasChainFilterer) ParseOwnershipTransferred(log types.Log) (*QGasChainOwnershipTransferred, error) {
	event := new(QGasChainOwnershipTransferred)
	if err := _QGasChain.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// QGasChainTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the QGasChain contract.
type QGasChainTransferIterator struct {
	Event *QGasChainTransfer // Event containing the contract specifics and raw log

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
func (it *QGasChainTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QGasChainTransfer)
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
		it.Event = new(QGasChainTransfer)
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
func (it *QGasChainTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QGasChainTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QGasChainTransfer represents a Transfer event raised by the QGasChain contract.
type QGasChainTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_QGasChain *QGasChainFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*QGasChainTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _QGasChain.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &QGasChainTransferIterator{contract: _QGasChain.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_QGasChain *QGasChainFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *QGasChainTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _QGasChain.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QGasChainTransfer)
				if err := _QGasChain.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_QGasChain *QGasChainFilterer) ParseTransfer(log types.Log) (*QGasChainTransfer, error) {
	event := new(QGasChainTransfer)
	if err := _QGasChain.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
