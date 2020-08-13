package grpcApi

import (
	"context"
	"go.uber.org/zap"
	"github.com/qlcchain/qlc-hub/wrapper"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/rpc/grpc/proto"
)

//WrapperAPI case
type WrapperAPI struct {
	logger 		*zap.SugaredLogger
	wap   		*wrapper.WrapperServer
}

//NewWrapperAPI api init
func NewWrapperAPI(wap *wrapper.WrapperServer) *WrapperAPI {
	wa := &WrapperAPI{
		wap:   		wap,
		logger: 	log.NewLogger("WrapperServerAPI"),
	}
	return wa
}

func (wa *WrapperAPI)Online(ctx context.Context,param *proto.OnlineRequest) (*proto.OnlineResponse, error) {
	wa.logger.Debugf("Exec Online request %+v", param)
	neoaccount,neocontract,ethaccount,ethcontract,activetime := wa.wap.WrapperOnline()
	return &proto.OnlineResponse{
		NeoAccount: 	neoaccount,
		EthAccount: 	ethaccount,
		NeoContract:	neocontract,
		EthContract:	ethcontract,
		ActiveTime:     activetime,
	}, nil
}

func (wa *WrapperAPI)EventStatCheck(ctx context.Context,param *proto.EventStatCheckRequest) (*proto.EventStatCheckResponse, error) {
	wa.logger.Debugf("Exec EventStatCheck request %+v", param)
	event,err := wa.wap.WrapperEventGetByTxhash(param.GetType(),param.GetHash())
	if err != nil {
		return nil,err
	}
	return &proto.EventStatCheckResponse{
		Type: 	param.GetType(),
		Hash: 	param.GetHash(),
		Status:	event.Status,
		Errno:	event.Errno,
	}, nil
}

func (wa *WrapperAPI)Nep5LockNotice(ctx context.Context,param *proto.Nep5LockNoticeRequest) (*proto.Nep5LockNoticeResponse, error) {
	wa.logger.Debugf("Exec Nep5LockNotice request %+v", param)
	result := wa.wap.WrapperNep5LockNotice(param.GetType(),param.GetAmount(),param.GetTxHash(),param.GetHash())
	return &proto.Nep5LockNoticeResponse{
		Type: 	param.GetType(),
		Hash: 	param.GetHash(),
		Result:	result,
	}, nil
}