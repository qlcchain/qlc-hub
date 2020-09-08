package main

import (
	"fmt"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/jessevdk/go-flags"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

var (
	version = "dev"
	date    = ""
	commit  = ""
)

var cfg = &config.Config{}

func main() {
	fmt.Println(logo())
	fmt.Printf("qlc-hub %s-%s.%s", version, commit, date)
	fmt.Println()

	if _, err := flag.ParseArgs(cfg, os.Args); err != nil {
		code := 1
		if fe, ok := err.(*flag.Error); ok {
			if fe.Type == flag.ErrHelp {
				code = 0
			}
		}
		fmt.Println(err)
		os.Exit(code)
	}

	if err := cfg.Verify(); err != nil {
		fmt.Println(util.ToIndentString(cfg))
		log.Root.Fatal(err)
	}

	if cfg.Verbose {
		cfg.LogLevel = "debug"
	}
	log.Setup(cfg.LogDir(), cfg.LogLevel)

	logger := log.NewLogger("main")
	logger.Info(util.ToIndentString(cfg))

	for i := 0; i < 10; i++ {
		if token, err := cfg.JwtManager.Generate(jwt.User); err == nil {
			logger.Infof("%d: %s", i, token)
		}
	}
	server := grpc.NewServer(cfg)
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
   ____  _      _____   _    _ _    _ ____  
  / __ \| |    / ____| | |  | | |  | |  _ \ 
 | |  | | |   | |      | |__| | |  | | |_) |
 | |  | | |   | |      |  __  | |  | |  _ < 
 | |__| | |___| |____  | |  | | |__| | |_) |
  \___\_\______\_____| |_|  |_|\____/|____/ 
`
}
