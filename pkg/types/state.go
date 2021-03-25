package types

import "strings"

//go:generate msgp
type SwapState byte

const (
	// deposit
	DepositPending SwapState = iota
	DepositDone

	// withdraw
	WithDrawPending
	WithDrawDone

	WithDrawFail
	DepositRefund
	Invalid
)

func SwapStateToString(t SwapState) string {
	switch t {
	case DepositPending:
		return "DepositPending"
	case DepositDone:
		return "DepositDone"
	case DepositRefund:
		return "DepositRefund"
	case WithDrawPending:
		return "WithDrawPending"
	case WithDrawDone:
		return "WithDrawDone"
	case WithDrawFail:
		return "WithDrawFail"
	default:
		return "Invalid"
	}
}

func StringToSwapState(t string) SwapState {
	switch t {
	case "DepositPending":
		return DepositPending
	case "DepositDone":
		return DepositDone
	case "WithDrawPending":
		return WithDrawPending
	case "WithDrawDone":
		return WithDrawDone
	case "WithDrawFail":
		return WithDrawFail
	default:
		return Invalid
	}
}

type ChainType byte

const (
	ETH ChainType = iota
	NEO
	BSC
	InvalidChain
)

func StringToChainType(s string) ChainType {
	t := strings.ToUpper(s)
	switch t {
	case "ETH":
		return ETH
	case "NEO":
		return NEO
	case "BSC":
		return BSC
	default:
		return InvalidChain
	}
}

func ChainTypeToString(c ChainType) string {
	switch c {
	case ETH:
		return "ETH"
	case NEO:
		return "NEO"
	case BSC:
		return "BSC"
	default:
		return "InvalidChain"
	}
}

type SwapType byte

const (
	Deposit SwapType = iota
	Withdraw
)
