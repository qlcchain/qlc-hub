package signer

import (
	"context"
	"time"

	"google.golang.org/grpc/backoff"

	"github.com/qlcchain/qlc-hub/pkg/util"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
)

type SignerClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      *grpc.ClientConn
	accessToken string
	logger      *zap.SugaredLogger
	timeout     time.Duration
}

func NewSigner(cfg *config.Config) (*SignerClient, error) {
	_, host, err := util.Scheme(cfg.SignerEndPoint)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	i := &SignerClient{
		ctx:         ctx,
		cancel:      cancel,
		accessToken: cfg.SignerToken,
		timeout:     5 * time.Second,
		logger:      log.NewLogger("signer/client"),
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	cc, err := grpc.DialContext(timeout, host, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: time.Second * 3,
		}),
		grpc.WithUnaryInterceptor(i.unary()), grpc.WithStreamInterceptor(i.stream()))
	if err != nil {
		return nil, err
	}
	i.client = cc
	i.scheduleRefreshToken(time.Hour * 12)
	return i, nil
}

func (i *SignerClient) refreshToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()

	c := proto.NewTokenServiceClient(i.client)
	if req, err := c.Refresh(ctx, &proto.RefreshRequest{Token: i.accessToken}); err == nil {
		i.accessToken = req.Token
	} else {
		return err
	}
	return nil
}

func (i *SignerClient) Sign(t proto.SignType, address string, rawData []byte) (*proto.SignResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()

	c := proto.NewSignServiceClient(i.client)
	return c.Sign(ctx, &proto.SignRequest{
		Type:    t,
		Address: address,
		RawData: rawData,
	})
}

func (i *SignerClient) AddressList(t proto.SignType) (*proto.AddressResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()

	c := proto.NewTokenServiceClient(i.client)
	return c.AddressList(ctx, &proto.AddressRequest{
		Type: t,
	})
}

func (i *SignerClient) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := i.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			select {
			case <-i.ctx.Done():
				return
			default:
				time.Sleep(wait)
				if err := i.refreshToken(); err != nil {
					wait = time.Second
				} else {
					wait = refreshDuration
				}
			}
		}
	}()

	return nil
}

func (i *SignerClient) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", i.accessToken)
}

func (i *SignerClient) unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(i.attachToken(ctx), method, req, reply, cc, opts...)
	}
}

func (i *SignerClient) stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(i.attachToken(ctx), desc, cc, method, opts...)
	}
}

func (i *SignerClient) Stop() {
	i.cancel()
	i.client.Close()
}
