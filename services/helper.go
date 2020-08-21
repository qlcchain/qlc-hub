package services

import (
	"github.com/qlcchain/qlc-hub/services/context"
	"github.com/qlcchain/qlc-hub/wrapper"
)

//RegisterServices register services to chain context
func RegisterServices(cs *context.ServiceContext) error {
	cfgFile := cs.ConfigFile()

	logService := NewLogService(cfgFile)
	_ = cs.Register(context.LogService, logService)
	_ = logService.Init()

	if rpcService, err := NewRPCService(cfgFile); err != nil {
		return err
	} else {
		_ = cs.Register(context.RPCService, rpcService)
	}
	wrapper.WrapperSqlInit(cfgFile)
	w := wrapper.NewWrapperServer(cfgFile)
	w.WrapperEventInit()
	return nil
}
