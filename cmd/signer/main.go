package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/jessevdk/go-flags"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	version = "dev"
	date    = ""
	commit  = ""
)

var cfg = &config.SignerConfig{}

func main() {
	fmt.Println(logo())
	fmt.Printf("signer %s-%s.%s", version, commit, date)
	fmt.Println()

	if _, err := flag.ParseArgs(cfg, os.Args); err != nil {
		code := 1
		if fe, ok := err.(*flag.Error); ok {
			if fe.Type == flag.ErrHelp {
				code = 0
			} else {
				log.Root.Error(err)
			}
		}

		os.Exit(code)
	}

	if err := cfg.Verify(); err != nil {
		fmt.Println(util.ToIndentString(cfg))
		log.Root.Fatal(err)
	}

	if cfg.Verbose {
		cfg.LogLevel = "debug"
	}
	_ = log.Setup(cfg.LogDir(), cfg.LogLevel)

	logger := log.NewLogger("signer/main")
	logger.Info(util.ToIndentString(cfg))

	r1 := cfg.AddressList(pb.SignType_NEO)
	logger.Info("NEO: ", util.ToIndentString(r1))

	r2 := cfg.AddressList(pb.SignType_ETH)
	logger.Info("ETH: ", util.ToIndentString(r2))

	for i := 0; i < 10; i++ {
		if token, err := cfg.JwtManager.Generate(jwt.User); err == nil {
			logger.Debugf("%d: %s", i, token)
		}
	}

	server, err := grpc.NewSignerServer(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	if err := server.Start(); err != nil {
		logger.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	if server != nil {
		server.Stop()
	}
}

func logo() string {
	return `
 ____     ___   ___     ___ ______  __ __       _____ ____   ____  ____     ___  ____  
|    \   /  _] /   \   /  _]      ||  |  |     / ___/|    | /    ||    \   /  _]|    \ 
|  _  | /  [_ |     | /  [_|      ||  |  |    (   \_  |  | |   __||  _  | /  [_ |  D  )
|  |  ||    _]|  O  ||    _]_|  |_||  _  |     \__  | |  | |  |  ||  |  ||    _]|    / 
|  |  ||   [_ |     ||   [_  |  |  |  |  |     /  \ | |  | |  |_ ||  |  ||   [_ |    \ 
|  |  ||     ||     ||     | |  |  |  |  |     \    | |  | |     ||  |  ||     ||  .  \
|__|__||_____| \___/ |_____| |__|  |__|__|      \___||____||___,_||__|__||_____||__|\_|
                                                                                       
`
}
