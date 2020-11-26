package types

//go:generate msgp
type SwapState byte

const (
	// deposit
	DepositPending SwapState = iota
	DepositDone

	// withdraw
	WithDrawPending
	WithDrawDone

	Invalid
)

func SwapStateToString(t SwapState) string {
	switch t {
	case DepositPending:
		return "DepositPending"
	case DepositDone:
		return "DepositDone"
	case WithDrawPending:
		return "WithDrawPending"
	case WithDrawDone:
		return "WithDrawDone"
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
	default:
		return Invalid
	}
}

type ChainType byte

const (
	ETH ChainType = iota
	NEO
)
