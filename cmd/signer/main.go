package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	rpc "github.com/qlcchain/qlc-hub/grpc"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"

	flag "github.com/jessevdk/go-flags"

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
			}
		}
		log.Root.Error(err)
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

	logger := log.NewLogger("main")
	logger.Info(util.ToIndentString(cfg))

	r1 := cfg.AddressList(pb.SignType_NEO)
	logger.Info("NEO: ", util.ToIndentString(r1))

	r2 := cfg.AddressList(pb.SignType_ETH)
	logger.Info("NEO: ", util.ToIndentString(r2))

	server, err := NewServer(cfg)
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

type Server struct {
	srv    *grpc.Server
	cfg    *config.SignerConfig
	logger *zap.SugaredLogger
}

func NewServer(cfg *config.SignerConfig) (*Server, error) {
	i := rpc.NewAuthInterceptor(cfg.JwtManager)
	srv := grpc.NewServer(grpc.UnaryInterceptor(i.Unary()), grpc.StreamInterceptor(i.Stream()))

	return &Server{
		srv:    srv,
		cfg:    cfg,
		logger: log.NewLogger("signer/srv"),
	}, nil
}

func (s *Server) Start() error {
	network, address, err := util.Scheme(s.cfg.GRPCListenAddress)
	if err != nil {
		return err
	}

	lis, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	s.registerApi()
	s.logger.Debugf("server start at %s", address)
	reflection.Register(s.srv)
	return s.srv.Serve(lis)
}

func (s *Server) Stop() {
	s.srv.Stop()
}

func (s *Server) registerApi() {
	pb.RegisterSignServiceServer(s.srv, apis.NewSignerService(s.cfg))
	pb.RegisterTokenServiceServer(s.srv, apis.NewTokenService(s.cfg))
}
