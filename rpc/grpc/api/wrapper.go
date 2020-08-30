package api

import (
	"context"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/rpc/grpc/proto"
	"github.com/qlcchain/qlc-hub/wrapper"
	"go.uber.org/zap"
)

//WrapperAPI case
type WrapperAPI struct {
	logger *zap.SugaredLogger
	wap    *wrapper.WrapperServer
}

//NewWrapperAPI api init
func NewWrapperAPI() *WrapperAPI {
	wap := wrapper.NewWrapperServer("")
	wa := &WrapperAPI{
		wap:    wap,
		logger: log.NewLogger("WrapperServerAPI"),
	}
	return wa
}

//Online wrapper online check
func (wa *WrapperAPI) Online(ctx context.Context, param *proto.OnlineRequest) (*proto.OnlineResponse, error) {
	wa.logger.Debugf("Exec Online request %+v", param)
	neoaccount, neocontract, ethaccount, ethcontract, activetime := wa.wap.WrapperOnline()
	return &proto.OnlineResponse{
		NeoAccount:  neoaccount,
		EthAccount:  ethaccount,
		NeoContract: neocontract,
		EthContract: ethcontract,
		ActiveTime:  activetime,
	}, nil
}

//EventStatCheck wrapper event status check by lock_hash
func (wa *WrapperAPI) EventStatCheck(ctx context.Context, param *proto.EventStatCheckRequest) (*proto.EventStatCheckResponse, error) {
	//wa.logger.Debugf("Exec EventStatCheck request %+v", param)
	event, err := wa.wap.WrapperEventGetByLockhash(param.GetType(), param.GetHash())
	if err != nil {
		wa.logger.Debugf("get err(%s)", err)
		return &proto.EventStatCheckResponse{
			Type:   param.GetType(),
			Hash:   param.GetHash(),
			Status: -1,
			Errno:  wrapper.CchGetEventStatusRetNoTxhash,
		}, nil
	} else {
		//wa.logger.Debugf("get Status(%d)", event.Status)
		return &proto.EventStatCheckResponse{
			Type:   param.GetType(),
			Hash:   param.GetHash(),
			Status: event.Status,
			Errno:  event.Errno,
		}, nil
	}

}

//Nep5LockNotice nep5 lock notice by app user
func (wa *WrapperAPI) Nep5LockNotice(ctx context.Context, param *proto.Nep5LockNoticeRequest) (*proto.Nep5LockNoticeResponse, error) {
	wa.logger.Debugf("Exec Nep5LockNotice request %+v", param)
	result := wa.wap.WrapperNep5LockNotice(param.GetType(), param.GetAmount(), param.GetUserlocknum(), param.GetTxhash(), param.GetHash(), param.GetSource())
	return &proto.Nep5LockNoticeResponse{
		Type:   param.GetType(),
		Hash:   param.GetHash(),
		Result: result,
	}, nil
}

//EthIssueLock eth smartcontract IssueLock
func (wa *WrapperAPI) EthIssueLock(ctx context.Context, param *proto.EthIssueLockRequest) (*proto.EthIssueLockResponse, error) {
	wa.logger.Debugf("Exec EthIssueLock request %+v", param)
	result, txhash, err := wa.wap.WrapperEthIssueLock(param.GetAmount(), param.GetLockhash())
	if err != nil {
		wa.logger.Debugf("Exec EthIssueLock get ERR", err)
	}
	return &proto.EthIssueLockResponse{
		Result: result,
		Txhash: txhash,
	}, nil
}

//EthIssueFetch eth smartcontract IssueFetch
func (wa *WrapperAPI) EthIssueFetch(ctx context.Context, param *proto.EthIssueFetchRequest) (*proto.EthIssueFetchResponse, error) {
	wa.logger.Debugf("Exec EthIssueFetch request %+v", param)
	result, txhash, err := wa.wap.WrapperEthIssueFetch(param.GetLockhash())
	if err != nil {
		wa.logger.Debugf("Exec EthIssueFetch get ERR", err)
	}
	return &proto.EthIssueFetchResponse{
		Result: result,
		Txhash: txhash,
	}, nil
}

//EthDestoryUnlock eth smartcontract DestoryUnlock
func (wa *WrapperAPI) EthDestoryUnlock(ctx context.Context, param *proto.EthDestoryUnlockRequest) (*proto.EthDestoryUnlockResponse, error) {
	wa.logger.Debugf("Exec EthDestoryUnlock request %+v", param)
	result, txhash, err := wa.wap.WrapperEthDestoryUnlock(param.GetLockhash(), param.GetLocksource())
	if err != nil {
		wa.logger.Debugf("Exec EthDestoryUnlock get ERR", err)
	}
	return &proto.EthDestoryUnlockResponse{
		Result: result,
		Txhash: txhash,
	}, nil
}

//EthUcallerDestorylock eth smartcontract user caller Destorylock
func (wa *WrapperAPI) EthUcallerDestoryLock(ctx context.Context, param *proto.EthUcallerDestoryLockRequest) (*proto.EthUcallerDestoryLockResponse, error) {
	wa.logger.Debugf("Exec EthUcallerDestoryLock request %+v", param)
	result, txhash, err := wa.wap.WrapperEthUcallerDestoryLock(param.GetAmount(), param.GetLockhash())
	if err != nil {
		wa.logger.Debugf("Exec EthUcallerDestoryLock get ERR", err)
	}
	return &proto.EthUcallerDestoryLockResponse{
		Result: result,
		Txhash: txhash,
	}, nil
}

//EthGetTransationInfo eth smartcontract get blockinfo by txhash
func (wa *WrapperAPI) EthGetTransationInfo(ctx context.Context, param *proto.EthGetTransationInfoRequest) (*proto.EthGetTransationInfoResponse, error) {
	wa.logger.Debugf("Exec EthGetTransationInfo request %+v", param)
	result, info, err := wa.wap.WrapperEthGetTransationInfo(param.GetInfotype(), param.GetTxhash())
	if err != nil {
		wa.logger.Debugf("Exec EthGetTransationInfo get ERR", err)
	}
	return &proto.EthGetTransationInfoResponse{
		Result:   result,
		Infotype: param.GetInfotype(),
		Info:     info,
	}, nil
}

//EthGetAccountInfo eth smartcontract get account info  by address
func (wa *WrapperAPI) EthGetAccountInfo(ctx context.Context, param *proto.EthGetAccountInfoRequest) (*proto.EthGetAccountInfoResponse, error) {
	wa.logger.Debugf("Exec EthGetAccountInfo request %+v", param)
	result, info, err := wa.wap.WrapperEthGetAccountInfo(param.GetAddress())
	if err != nil {
		wa.logger.Debugf("Exec EthGetAccountInfo get ERR", err)
	}
	return &proto.EthGetAccountInfoResponse{
		Result: result,
		Info:   info,
	}, nil
}

//EthGetHashTimer eth smartcontract get hashtimer info  by lockhash
func (wa *WrapperAPI) EthGetHashTimer(ctx context.Context, param *proto.EthGetHashTimerRequest) (*proto.EthGetHashTimerResponse, error) {
	wa.logger.Debugf("Exec EthGetHashTimer request %+v", param)
	result, elog, err := wa.wap.WrapperEthGetHashTimer(param.GetLockhash())
	if err != nil {
		wa.logger.Debugf("Exec EthGetGetHashTimer get ERR", err)
	}
	return &proto.EthGetHashTimerResponse{
		Result:     result,
		Amount:     elog.Amount,
		Locknum:    elog.LockNum,
		Unlocknum:  elog.UnlockNum,
		Account:    elog.Account,
		Locksource: elog.LockSource,
	}, nil
}

//Nep5WrapperLock neo smartcontract lock by lockhash
func (wa *WrapperAPI) Nep5WrapperLock(ctx context.Context, param *proto.Nep5WrapperLockRequest) (*proto.Nep5WrapperLockResponse, error) {
	wa.logger.Debugf("Exec Nep5WrapperLock request %+v", param)
	result, txhash, msg, err := wa.wap.WrapperNep5WrapperLock(param.GetAmount(), param.GetBlocknum(), param.GetEthaddress(), param.GetLockhash())
	if err != nil {
		wa.logger.Debugf("Exec Nep5WrapperLock get ERR", err)
	}
	return &proto.Nep5WrapperLockResponse{
		Result: result,
		Txhash: txhash,
		Msg:    msg,
	}, nil
}

//Nep5WrapperUnlock neo smartcontract unlock by locksource
func (wa *WrapperAPI) Nep5WrapperUnlock(ctx context.Context, param *proto.Nep5WrapperUnlockRequest) (*proto.Nep5WrapperUnlockResponse, error) {
	wa.logger.Debugf("Exec Nep5WrapperUnlock request %+v", param)
	result, txhash, msg, err := wa.wap.WrapperNep5WrapperUnlock(param.GetEthaddress(), param.GetLocksource())
	if err != nil {
		wa.logger.Debugf("Exec Nep5WrapperUnlock get ERR", err)
	}
	return &proto.Nep5WrapperUnlockResponse{
		Result: result,
		Txhash: txhash,
		Msg:    msg,
	}, nil
}

//Nep5WrapperRefund neo smartcontract refound by locksource
func (wa *WrapperAPI) Nep5WrapperRefund(ctx context.Context, param *proto.Nep5WrapperRefundRequest) (*proto.Nep5WrapperRefundResponse, error) {
	wa.logger.Debugf("Exec Nep5WrapperRefund request %+v", param)
	result, txhash, msg, err := wa.wap.WrapperNep5WrapperRefund(param.GetLocksource())
	if err != nil {
		wa.logger.Debugf("Exec Nep5WrapperRefund get ERR", err)
	}
	return &proto.Nep5WrapperRefundResponse{
		Result: result,
		Txhash: txhash,
		Msg:    msg,
	}, nil
}

//Nep5GetTxInfo neo get txinfo by txid
func (wa *WrapperAPI) Nep5GetTxInfo(ctx context.Context, param *proto.Nep5GetTxInfoRequest) (*proto.Nep5GetTxInfoResponse, error) {
	wa.logger.Debugf("Exec Nep5GetTxInfo request %+v", param)
	result, action, fromaddr, toaddr, amount, err := wa.wap.WrapperNep5GetTxInfo(param.GetTxhash())
	if err != nil {
		wa.logger.Debugf("Exec Nep5GetTxInfo get ERR", err)
	}
	return &proto.Nep5GetTxInfoResponse{
		Result:   result,
		Action:   action,
		Fromaddr: fromaddr,
		Toaddr:   toaddr,
		Amount:   amount,
	}, nil
}
