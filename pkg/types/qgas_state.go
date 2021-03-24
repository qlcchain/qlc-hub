package types

//go:generate msgp
type QGasSwapState byte

const (
	// pledge
	QGasPledgeInit QGasSwapState = iota
	QGasPledgePending
	QGasPledgeDone

	// withdraw
	QGasWithDrawInit
	QGasWithDrawPending
	QGasWithDrawDone

	QGasInvalid
)

func QGasSwapStateToString(t QGasSwapState) string {
	switch t {
	case QGasPledgeInit:
		return "QGasPledgeInit"
	case QGasPledgePending:
		return "QGasPledgePending"
	case QGasPledgeDone:
		return "QGasPledgeDone"
	case QGasWithDrawPending:
		return "QGasWithDrawPending"
	case QGasWithDrawInit:
		return "QGasWithDrawInit"
	case QGasWithDrawDone:
		return "QGasWithDrawDone"
	default:
		return "QGasInvalid"
	}
}

func StringToQGasSwapState(t string) QGasSwapState {
	switch t {
	case "QGasPledgeInit":
		return QGasPledgeInit
	case "QGasPledgePending":
		return QGasPledgePending
	case "QGasPledgeDone":
		return QGasPledgeDone
	case "QGasWithDrawInit":
		return QGasWithDrawInit
	case "QGasWithDrawPending":
		return QGasWithDrawPending
	case "QGasWithDrawDone":
		return QGasWithDrawDone
	default:
		return QGasInvalid
	}
}

//
//type ChainType byte
//
//const (
//	ETH ChainType = iota
//	NEO
//)
//
type QGasSwapType byte

const (
	QGasDeposit QGasSwapType = iota
	QGasWithdraw
)
