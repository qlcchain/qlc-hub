package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	rpc    *grpc.Server
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config
	logger *zap.SugaredLogger
}

func NewServer(cfg *config.Config) *Server {
	gRpcServer := grpc.NewServer()
	ctx, cancel := context.WithCancel(context.Background())

	r := &Server{
		cfg:    cfg,
		rpc:    gRpcServer,
		ctx:    ctx,
		cancel: cancel,
		logger: log.NewLogger("rpc"),
	}
	return r
}

func (g *Server) Start() error {
	network, address, err := scheme(g.cfg.RPCCfg.GRPCListenAddress)
	if err != nil {
		return err
	}

	lis, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	if err := g.registerApi(); err != nil {
		g.logger.Error(err)
		return fmt.Errorf("registerApi: %s", err)
	}
	reflection.Register(g.rpc)
	go func() {
		if err := g.rpc.Serve(lis); err != nil {
			g.logger.Error(err)
		}
	}()
	go func() {
		if err := g.newGateway(address, g.cfg.RPCCfg.ListenAddress); err != nil {
			g.logger.Errorf("start gateway: %s", err)
		}
	}()
	return nil
}

func (g *Server) newGateway(grpcAddress, gwAddress string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	// no need proxy for internal gateway to internal grpc server
	optDial := grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		network := "tcp"
		g.logger.Debugf("WithContextDialer addr %s", addr)
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	})
	opts := []grpc.DialOption{grpc.WithInsecure(), optDial}
	if err := registerGWApi(ctx, gwmux, grpcAddress, opts); err != nil {
		return fmt.Errorf("gateway register: %s", err)
	}
	_, address, err := scheme(gwAddress)
	if err != nil {
		return err
	}
	if err := http.ListenAndServe(address, gwmux); err != nil {
		g.logger.Error(err)
	}
	return nil
}

func (g *Server) Stop() {
	g.rpc.Stop()
}

func scheme(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}

func (g *Server) registerApi() error {
	eth, err := apis.NewEthAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterEthAPIServer(g.rpc, eth)
	neo, err := apis.NewNeoAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterNeoAPIServer(g.rpc, neo)
	event, err := apis.NewEventAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterEventAPIServer(g.rpc, event)
	debug, err := apis.NewDebugAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterDebugAPIServer(g.rpc, debug)
	return nil
}

func registerGWApi(ctx context.Context, gwmux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := pb.RegisterEthAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterNeoAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterDebugAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterEventAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	return nil
}
