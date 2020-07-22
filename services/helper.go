package services

import (
	"github.com/qlcchain/qlc-hub/services/context"
)

//RegisterServices register services to chain context
func RegisterServices(cs *context.ServiceContext) error {
	cfgFile := cs.ConfigFile()

	cfg, err := cs.Config()
	if err != nil {
		return err
	}

	logService := NewLogService(cfgFile)
	_ = cs.Register(context.LogService, logService)
	_ = logService.Init()

	if cfg.P2P.IsBootNode {
		httpService := NewHttpService(cfgFile)
		_ = cs.Register(context.BootNodeHttpService, httpService)
	}

	if len(cfg.P2P.BootNodes) > 0 {
		netService, err := NewP2PService(cfgFile)
		if err != nil {
			return err
		}
		_ = cs.Register(context.P2PService, netService)
	}

	return nil
}
