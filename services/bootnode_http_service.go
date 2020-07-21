/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/common"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/services/context"
)

type HttpService struct {
	common.ServiceLifecycle
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func NewHttpService(cfgFile string) *HttpService {
	cc := context.NewServiceContext(cfgFile)
	cfg, _ := cc.Config()
	return &HttpService{cfg: cfg, logger: log.NewLogger("http_service")}
}

func (hs *HttpService) Init() error {
	if !hs.PreInit() {
		return errors.New("pre init fail")
	}
	defer hs.PostInit()
	return nil
}

func (hs *HttpService) Start() error {
	if !hs.PreStart() {
		return errors.New("pre start fail")
	}
	defer hs.PostStart()

	http.HandleFunc("/bootNode", func(w http.ResponseWriter, r *http.Request) {
		rl := strings.ReplaceAll(hs.cfg.P2P.Listen, "0.0.0.0", hs.cfg.P2P.ListeningIp)
		bootNode := rl + "/p2p/" + hs.cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe(hs.cfg.P2P.BootNodeHttpServer, nil); err != nil {
			hs.logger.Error(err)
		}
	}()
	return nil
}

func (hs *HttpService) Stop() error {
	if !hs.PreStop() {
		return errors.New("pre stop fail")
	}
	defer hs.PostStop()
	return nil
}

func (hs *HttpService) Status() int32 {
	return hs.State()
}
