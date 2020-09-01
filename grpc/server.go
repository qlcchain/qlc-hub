package grpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
)

type Server struct {
	rpc    *grpc.Server
	srv    *http.Server
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config
	eth    *ethclient.Client
	neo    *neo.Transaction
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
	if err := g.checkBaseInfo(); err != nil {
		return err
	}

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
			g.logger.Errorf("gateway listen err: %s", err)
		}
	}()
	return nil
}

func (g *Server) checkBaseInfo() error {
	eClient, err := ethclient.Dial(g.cfg.EthereumCfg.EndPoint)
	if err != nil {
		return fmt.Errorf("eth dail: %s", err)
	}
	if _, err := eClient.BlockByNumber(context.Background(), nil); err != nil {
		return fmt.Errorf("eth node connect timeout: %s", err)
	}
	g.eth = eClient

	nTransaction, err := neo.NewTransaction(g.cfg.NEOCfg.EndPoint, g.cfg.NEOCfg.Contract)
	if err != nil {
		return fmt.Errorf("neo transaction: %s", err)
	}
	if err := nTransaction.Client().Ping(); err != nil {
		return fmt.Errorf("neo node connect timeout: %s", err)
	}
	g.neo = nTransaction

	_, err = wallet.NewAccountFromWIF(g.cfg.NEOCfg.WIF)
	if err != nil {
		return fmt.Errorf("invalid nep5 wif: %s [%s]", err, g.cfg.NEOCfg.WIF)
	}
	_, _, err = eth.GetAccountByPriKey(g.cfg.EthereumCfg.Account)
	if err != nil {
		return fmt.Errorf("invalid erc20 account: %s [%s]", err, g.cfg.EthereumCfg.Account)
	}
	return nil
}

func (g *Server) newGateway(grpcAddress, gwAddress string) error {
	ctx, cancel := context.WithCancel(context.Background())
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
	g.srv = &http.Server{
		Addr:    address,
		Handler: gwmux,
	}

	g.srv.RegisterOnShutdown(func() {
		g.logger.Debug("RESEful server shutdown")
	})

	if err := g.srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (g *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if g.srv != nil {
		if err := g.srv.Shutdown(ctx); err != nil {
			g.logger.Errorf("RESTful server shutdown failed:%+v", err)
		}
	}
	g.eth.Close()
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
	eth, err := apis.NewDepositAPI(g.ctx, g.cfg, g.neo, g.eth)
	if err != nil {
		return err
	}
	pb.RegisterDepositAPIServer(g.rpc, eth)
	neo, err := apis.NewWithdrawAPI(g.ctx, g.cfg, g.neo, g.eth)
	if err != nil {
		return err
	}
	pb.RegisterWithdrawAPIServer(g.rpc, neo)
	event, err := apis.NewEventAPI(g.ctx, g.cfg, g.neo, g.eth)
	if err != nil {
		return err
	}
	pb.RegisterEventAPIServer(g.rpc, event)
	debug, err := apis.NewDebugAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterDebugAPIServer(g.rpc, debug)
	info, err := apis.NewInfoAPI(g.ctx, g.cfg)
	if err != nil {
		return err
	}
	pb.RegisterInfoAPIServer(g.rpc, info)
	return nil
}

func registerGWApi(ctx context.Context, gwmux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := pb.RegisterDepositAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterWithdrawAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterInfoAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
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
