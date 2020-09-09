package jwt

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/qlcchain/qlc-hub/pkg/log"
)

type AuthInterceptor struct {
	logger     *zap.SugaredLogger
	Authorizer AuthorizeFn
}

type AuthorizeFn func(ctx context.Context, method string) error

func NewAuthInterceptor(Authorizer AuthorizeFn) *AuthInterceptor {
	return &AuthInterceptor{
		Authorizer: Authorizer,
		logger:     log.NewLogger("auth/interceptor"),
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		i.logger.Debug(info.FullMethod)
		if err := i.Authorizer(ctx, info.FullMethod); err != nil {
			i.logger.Errorf("%s: %s", info.FullMethod, err)
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		i.logger.Debug(info.FullMethod)
		if err := i.Authorizer(stream.Context(), info.FullMethod); err != nil {
			i.logger.Errorf("%s: %s", info.FullMethod, err)
			return err
		}
		return handler(srv, stream)
	}
}

func DefaultAuthorizer(jwtManager *JWTManager, accessibleRoles map[string][]string) AuthorizeFn {
	return func(ctx context.Context, method string) error {
		accessibleRoles, ok := accessibleRoles[method]
		if !ok {
			// everyone can access
			return nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := values[0]
		if accessToken == "" {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}
		claims, err := jwtManager.Verify(accessToken)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "access token is invalid: %v, %s", err, accessToken)
		}

		for _, role := range accessibleRoles {
			if claims.IsAuthorized(role) {
				return nil
			}
		}

		return status.Error(codes.PermissionDenied, "no permission to access this RPC")
	}
}
