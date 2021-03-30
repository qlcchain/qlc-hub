package grpc

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/apis"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/jwt"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/qlc"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

type Server struct {
	rpc     *grpc.Server
	srv     *http.Server
	signer  *signer.SignerClient
	ctx     context.Context
	cancel  context.CancelFunc
	cfg     *config.Config
	ethNep5 *eth.Transaction
	ethQGas *eth.Transaction
	bscNep5 *eth.Transaction
	bscQGas *eth.Transaction
	neo     *neo.Transaction
	qlc     *qlc.Transaction
	store   *gorm.DB
	logger  *zap.SugaredLogger
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

	g.logger.Info("grpc server started")

	return nil
}

func (g *Server) checkBaseInfo() error {
	signer, err := signer.NewSigner(g.cfg)
	if err != nil {
		return fmt.Errorf("new signer: %s", err)
	}

	if _, err := signer.Sign(pb.SignType_ETH, g.cfg.EthCfg.EthNep5Owner, bytes.Repeat([]byte{0}, 32)); err != nil {
		return fmt.Errorf("sign: %s", err)
	}
	g.signer = signer
	g.logger.Info("signer client connected successfully")

	ethNep5Transaction, err := eth.NewTransaction(g.cfg.EthCfg.EndPoints, g.cfg.EthCfg.EthNep5Contract)
	if err != nil {
		return fmt.Errorf("eth client: %s", err)
	}
	g.ethNep5 = ethNep5Transaction
	g.logger.Info("eth nep5 client connected successfully")

	ethQGasTransaction, err := eth.NewTransaction(g.cfg.EthCfg.EndPoints, g.cfg.EthCfg.EthQGasContract)
	if err != nil {
		return fmt.Errorf("eth client: %s", err)
	}
	g.ethQGas = ethQGasTransaction
	g.logger.Info("eth qgas client connected successfully")

	bscNep5Transaction, err := eth.NewTransaction(g.cfg.BscCfg.EndPoints, g.cfg.BscCfg.BscNep5Contract)
	if err != nil {
		return fmt.Errorf("eth client: %s", err)
	}
	g.bscNep5 = bscNep5Transaction
	g.logger.Info("bsc nep5 client connected successfully")

	bscQGasTransaction, err := eth.NewTransaction(g.cfg.BscCfg.EndPoints, g.cfg.BscCfg.BscQGasContract)
	if err != nil {
		return fmt.Errorf("eth client: %s", err)
	}
	g.bscQGas = bscQGasTransaction
	g.logger.Info("bsc qgas client connected successfully")

	nTransaction, err := neo.NewTransaction(g.cfg.NEOCfg.EndPoints, g.cfg.NEOCfg.Contract, signer)
	if err != nil {
		return fmt.Errorf("neo transaction: %s", err)
	}
	if c := nTransaction.Client(); c == nil {
		return fmt.Errorf("neo node connect timeout: %s", err)
	}

	g.neo = nTransaction
	g.logger.Info("neo client connected successfully")

	qTransaction, err := qlc.NewTransaction(g.cfg.QlcCfg.EndPoint, signer)
	if err != nil {
		return fmt.Errorf("qlc transaction: %s", err)
	}
	g.qlc = qTransaction
	g.logger.Info("qlc client connected successfully")

	//store, err := store.NewStore(g.cfg.DataDir())
	//if err != nil {
	//	return fmt.Errorf("new store fail: %s", err)
	//}
	db, err := db.NewDB(g.cfg.Database())
	if err != nil {
		return fmt.Errorf("new store fail: %s", err)
	}
	g.logger.Infof("store dir: %s ", g.cfg.Database())
	g.store = db

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
	handler := newCorsHandler(gwmux, g.cfg.RPCCfg.CORSAllowedOrigins)

	g.srv = &http.Server{
		Addr:    address,
		Handler: handler,
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
	g.ethNep5.Client().Close()
	g.ethQGas.Client().Close()
	g.bscNep5.Client().Close()
	g.bscQGas.Client().Close()
	g.rpc.Stop()
	g.qlc.Client().Close()
	g.signer.Stop()

	g.logger.Info("grpc stopped")
}

func (g *Server) registerApi() error {
	pb.RegisterDepositAPIServer(g.rpc, apis.NewDepositAPI(g.ctx, g.cfg, g.neo, g.ethNep5, g.signer, g.store))
	pb.RegisterWithdrawAPIServer(g.rpc, apis.NewWithdrawAPI(g.ctx, g.cfg, g.neo, g.ethNep5, g.store))
	pb.RegisterInfoAPIServer(g.rpc, apis.NewInfoAPI(g.ctx, g.cfg, g.neo, g.ethNep5, g.ethQGas, g.bscQGas, g.store))
	pb.RegisterDebugAPIServer(g.rpc, apis.NewDebugAPI(g.ctx, g.cfg, g.ethNep5, g.neo, g.store))
	pb.RegisterQGasSwapAPIServer(g.rpc, apis.NewQGasSwapAPI(g.ctx, g.cfg, g.qlc, g.ethQGas, g.bscQGas, g.signer, g.store))
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
	if err := pb.RegisterQGasSwapAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	return nil
}

func authorizer(manager *jwt.JWTManager) jwt.AuthorizeFn {
	authorizer := jwt.DefaultAuthorizer(manager, map[string][]string{
		"/proto.DepositAPI/PackNeoTransaction":       jwt.Both,
		"/proto.DepositAPI/SendNeoTransaction":       jwt.Both,
		"/proto.DepositAPI/NeoTransactionConfirmed":  jwt.Both,
		"/proto.DepositAPI/EthTransactionSent":       jwt.Both,
		"/proto.DepositAPI/GetEthOwnerSign":          jwt.Both,
		"/proto.DepositAPI/Refund":                   jwt.Both,
		"/proto.DepositAPI/EthTransactionID":         jwt.Both,
		"/proto.WithdrawAPI/EthTransactionConfirmed": jwt.Both,
		"/proto.WithdrawAPI/EthTransactionSent":      jwt.Both,
		"/proto.InfoAPI/Ping":                        jwt.Both,
		"/proto.InfoAPI/Config":                      jwt.Both,
		"/proto.InfoAPI/SwapInfoList":                jwt.Both,
		"/proto.InfoAPI/SwapInfosByAddress":          jwt.Both,
		"/proto.InfoAPI/SwapInfoByTxHash":            jwt.Both,
		"/proto.InfoAPI/SwapInfosByState":            jwt.Both,
		"/proto.InfoAPI/SwapCountByState":            jwt.Both,
		"/proto.InfoAPI/SwapAmountByState":           jwt.Both,
		"/proto.InfoAPI/SwapAmountByAddress":         jwt.Both,
		"/proto.InfoAPI/CheckNeoTransaction":         jwt.Both,
		"/proto.InfoAPI/CheckEthTransaction":         jwt.Both,
		"/proto.QGasSwapAPI/GetPledgeSendBlock":      jwt.Both,
		"/proto.QGasSwapAPI/PledgeEthTxSent":         jwt.Both,
		"/proto.QGasSwapAPI/GetOwnerSign":            jwt.Both,
		"/proto.QGasSwapAPI/GetWithdrawRewardBlock":  jwt.Both,
		"/proto.QGasSwapAPI/WithdrawEthTxSent":       jwt.Both,
		"/proto.QGasSwapAPI/ProcessBlock":            jwt.Both,
		"/proto.QGasSwapAPI/SwapInfoList":            jwt.Both,
		"/proto.QGasSwapAPI/SwapInfoByTxHash":        jwt.Both,
		"/proto.QGasSwapAPI/SwapInfosByAddress":      jwt.Both,
		"/proto.QGasSwapAPI/SwapInfosByState":        jwt.Both,
		"/proto.QGasSwapAPI/SwapInfosCount":          jwt.Both,
		"/proto.QGasSwapAPI/SwapInfosAmount":         jwt.Both,
	})
	return authorizer
}

func newCorsHandler(srv http.Handler, allowedOrigins []string) http.Handler {
	if len(allowedOrigins) == 0 {
		return srv
	}
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
		MaxAge:         600,
		AllowedHeaders: []string{"*"},
	})
	return c.Handler(srv)
}
