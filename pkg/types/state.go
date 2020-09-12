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
	DepositNeoUnLockedPending //5
	DepositNeoUnLockedDone
	DepositEthFetchPending
	DepositEthFetchDone
	DepositNeoFetchPending
	DepositNeoFetchDone //10

	// withdraw
	WithDrawInit
	WithDrawEthLockedDone
	WithDrawNeoLockedPending
	WithDrawNeoLockedDone
	WithDrawNeoUnLockedPending //15
	WithDrawNeoUnLockedDone
	WithDrawEthUnlockPending
	WithDrawEthUnlockDone
	WithDrawNeoFetchPending
	WithDrawNeoFetchDone
	WithDrawEthFetchDone

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
	case DepositNeoFetchPending:
		return "DepositNeoFetchPending"
	case DepositNeoFetchDone:
		return "DepositNeoFetchDone"
	case WithDrawInit:
		return "WithDrawInit"
	case WithDrawEthLockedDone:
		return "WithDrawEthLockedDone"
	case WithDrawNeoLockedPending:
		return "WithDrawNeoLockedPending"
	case WithDrawNeoLockedDone:
		return "WithDrawNeoLockedDone"
	case WithDrawNeoUnLockedPending:
		return "WithDrawNeoUnLockedPending"
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
	default:
		return "Invalid"
	}
}
