package types

//go:generate msgp
type LockerState byte

const (
	// deposit
	DepositInit LockerState = iota
	DepositNeoLockedDone
	DepositEthLockedPending
	DepositEthLockedDone
	DepositEthUnLockedDone
	DepositNeoUnLockedPending
	DepositNeoUnLockedDone
	DepositEthFetchPending
	DepositEthFetchDone
	DepositNeoFetchPending
	DepositNeoFetchDone

	// withdraw
	WithDrawEthLockedDone = iota
	WithDrawNeoLockedPending
	WithDrawNeoLockedDone
	WithDrawNeoUnLockedDone
	WithDrawEthUnlockPending
	WithDrawEthUnlockDone
	WithDrawNeoFetchPending
	WithDrawNeoFetchDone
	WithDrawEthFetchDone

	Failed
	Invalid
)

func LockerStateToString(t LockerState) string {
	switch t {
	case DepositInit:
		return "DepositInit"
	case DepositNeoLockedDone:
		return "DepositNeoLockedDone"
	case DepositEthLockedPending:
		return "DepositEthLockedPending"
	case DepositEthLockedDone:
		return "DepositEthLockedDone"
	case DepositEthUnLockedDone:
		return "DepositEthUnLockedDone"
	case DepositNeoUnLockedPending:
		return "DepositNeoUnLockedPending"
	case DepositNeoUnLockedDone:
		return "DepositNeoUnLockedDone"
	case DepositEthFetchPending:
		return "DepositEthFetchPending"
	case DepositEthFetchDone:
		return "DepositEthFetchDone"
	case DepositNeoFetchDone:
		return "DepositNeoFetchDone"
	case WithDrawEthLockedDone:
		return "WithDrawEthLockedDone"
	case WithDrawNeoLockedPending:
		return "WithDrawNeoLockedPending"
	case WithDrawNeoLockedDone:
		return "WithDrawNeoLockedDone"
	case WithDrawNeoUnLockedDone:
		return "WithDrawNeoUnLockedDone"
	case WithDrawEthUnlockPending:
		return "WithDrawEthUnlockPending"
	case WithDrawEthUnlockDone:
		return "WithDrawEthUnlockDone"
	case WithDrawNeoFetchPending:
		return "WithDrawNeoFetchPending"
	case WithDrawNeoFetchDone:
		return "WithDrawNeoFetchDone"
	case WithDrawEthFetchDone:
		return "WithDrawEthFetchDone"
	case Failed:
		return "Failed"
	default:
		return "Invalid"
	}
}
