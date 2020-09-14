package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/util"
)

type SignerServer struct {
	srv    *grpc.Server
	cfg    *config.SignerConfig
	logger *zap.SugaredLogger
}

func NewSignerServer(cfg *config.SignerConfig) (*SignerServer, error) {
	authorizer := jwt.DefaultAuthorizer(cfg.JwtManager, map[string][]string{
		"/proto.TokenService/Refresh":     jwt.Both,
		"/proto.TokenService/AddressList": jwt.Both,
		"/proto.SignService/Sign":         jwt.Both,
	})

	i := jwt.NewAuthInterceptor(authorizer)
	srv := grpc.NewServer(grpc.UnaryInterceptor(i.Unary()), grpc.StreamInterceptor(i.Stream()))

	return &SignerServer{
		srv:    srv,
		cfg:    cfg,
		logger: log.NewLogger("signer/srv"),
	}, nil
}

func (s *SignerServer) Start() error {
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

func (s *SignerServer) Stop() {
	s.srv.Stop()
}

func (s *SignerServer) registerApi() {
	pb.RegisterSignServiceServer(s.srv, apis.NewSignerService(s.cfg))
	pb.RegisterTokenServiceServer(s.srv, apis.NewTokenService(s.cfg))
}
