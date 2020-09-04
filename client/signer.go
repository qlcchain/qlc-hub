package client

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
)

type AuthClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	client      *grpc.ClientConn
	accessToken string
	logger      *zap.SugaredLogger
}

func NewAuthClient(cfg *config.Config) (*AuthClient, error) {
	cc, err := grpc.Dial(cfg.SignerEndPoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	i := &AuthClient{
		ctx:         ctx,
		cancel:      cancel,
		client:      cc,
		accessToken: cfg.SignerToken,
		logger:      log.NewLogger("client/interceptor"),
	}
	i.scheduleRefreshToken(time.Hour * 12)
	return i, nil
}

func (i *AuthClient) refreshToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := proto.NewTokenServiceClient(i.client)
	if req, err := c.Refresh(ctx, &proto.RefreshRequest{Token: i.accessToken}); err == nil {
		i.accessToken = req.Token
	} else {
		return err
	}
	return nil
}

func (i *AuthClient) Sign(t proto.SignType, address string, rawData []byte) (*proto.SignResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := proto.NewSignServiceClient(i.client)
	if sign, err := c.Sign(ctx, &proto.SignRequest{
		Type:    t,
		Address: address,
		RawData: rawData,
	}); err == nil {
		return sign, nil
	} else {
		return nil, err
	}
}

func (i *AuthClient) scheduleRefreshToken(refreshDuration time.Duration) error {
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
				err := i.refreshToken()
				if err != nil {
					wait = time.Second
				} else {
					wait = refreshDuration
				}
			}
		}
	}()

	return nil
}

func (i *AuthClient) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", i.accessToken)
}

func (i *AuthClient) Unary() grpc.UnaryClientInterceptor {
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

func (i *AuthClient) Stream() grpc.StreamClientInterceptor {
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

func (i *AuthClient) Stop() {
	i.cancel()
	i.client.Close()
}
