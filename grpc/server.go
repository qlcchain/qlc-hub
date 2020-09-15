package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
)

type Server struct {
	rpc    *grpc.Server
	srv    *http.Server
	signer *signer.SignerClient
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config
	eth    *eth.Transaction
	neo    *neo.Transaction
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewServer(cfg *config.Config) *Server {
	authorizer := authorizer(cfg.JwtManager)
	i := jwt.NewAuthInterceptor(authorizer)
	gRpcServer := grpc.NewServer(grpc.UnaryInterceptor(i.Unary()), grpc.StreamInterceptor(i.Stream()))

	//gRpcServer := grpc.NewServer()
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		cfg:    cfg,
		rpc:    gRpcServer,
		ctx:    ctx,
		cancel: cancel,
		logger: log.NewLogger("rpc"),
	}
}

func (g *Server) Start() error {
	if err := g.checkBaseInfo(); err != nil {
		g.logger.Error(err)
		return err
	}
	g.logger.Info("config info checked")

	network, address, err := util.Scheme(g.cfg.RPCCfg.GRPCListenAddress)
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
	signer, err := signer.NewSigner(g.cfg)
	if err != nil {
		return err
	}
	g.signer = signer
	g.logger.Info("signer client connected successfully")

	eClient, err := ethclient.Dial(g.cfg.EthereumCfg.EndPoint)
	if err != nil {
		return fmt.Errorf("eth dail: %s", err)
	}
	if _, err := eClient.BlockByNumber(context.Background(), nil); err != nil {
		return fmt.Errorf("eth node connect timeout: %s", err)
	}
	eTransaction := eth.NewTransaction(eClient, signer, g.ctx, g.cfg.EthereumCfg.GasEndPoint, g.cfg.EthereumCfg.Contract)
	g.eth = eTransaction
	g.logger.Info("eth client connected successfully")

	nTransaction, err := neo.NewTransaction(g.cfg.NEOCfg.EndPoint, g.cfg.NEOCfg.Contract, signer)
	if err != nil {
		return fmt.Errorf("neo transaction: %s", err)
	}
	if err := nTransaction.Client().Ping(); err != nil {
		return fmt.Errorf("neo node connect timeout: %s", err)
	}
	g.neo = nTransaction
	g.logger.Info("neo client connected successfully")

	store, err := store.NewStore(g.cfg.DataDir())
	if err != nil {
		return fmt.Errorf("new store fail: %s", err)
	}
	g.logger.Infof("store dir: %s ", g.cfg.DataDir())
	g.store = store

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
	_, address, err := util.Scheme(gwAddress)
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
	g.eth.Client().Close()
	g.rpc.Stop()
	g.signer.Stop()

	g.store.Close() //todo wait all server stop

}

func (g *Server) registerApi() error {
	pb.RegisterDepositAPIServer(g.rpc, apis.NewDepositAPI(g.ctx, g.cfg, g.neo, g.eth, g.store))
	pb.RegisterWithdrawAPIServer(g.rpc, apis.NewWithdrawAPI(g.ctx, g.cfg, g.neo, g.eth, g.store))
	pb.RegisterEventAPIServer(g.rpc, apis.NewEventAPI(g.ctx, g.cfg, g.neo, g.eth, g.store))
	pb.RegisterDebugAPIServer(g.rpc, apis.NewDebugAPI(g.ctx, g.cfg, g.eth, g.neo, g.store))
	pb.RegisterInfoAPIServer(g.rpc, apis.NewInfoAPI(g.ctx, g.cfg, g.neo, g.eth, g.store))
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

func authorizer(manager *jwt.JWTManager) jwt.AuthorizeFn {
	authorizer := jwt.DefaultAuthorizer(manager, map[string][]string{
		"/proto.DepositAPI/Lock":                    jwt.Both,
		"/proto.DepositAPI/Fetch":                   jwt.Both,
		"/proto.WithdrawAPI/Lock":                   jwt.Both,
		"/proto.WithdrawAPI/Claim":                  jwt.Both,
		"/proto.EventAPI/Event":                     jwt.Both,
		"/proto.InfoAPI/Ping":                       jwt.Admin,
		"/proto.DebugAPI/HashTimer":                 jwt.Both,
		"/proto.DebugAPI/LockerInfosCount":          jwt.Both,
		"/proto.DebugAPI/InterruptLocker":           jwt.Admin,
		"/proto.DebugAPI/DeleteLockerInfo":          jwt.Admin,
		"/proto.DebugAPI/LockerInfosByDeletedState": jwt.Admin,
	})
	return authorizer
}
