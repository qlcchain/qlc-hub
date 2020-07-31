package services

import (
	"github.com/qlcchain/qlc-hub/services/context"
)

//RegisterServices register services to chain context
func RegisterServices(cs *context.ServiceContext) error {
	cfgFile := cs.ConfigFile()

	logService := NewLogService(cfgFile)
	_ = cs.Register(context.LogService, logService)
	_ = logService.Init()

	return nil
}
