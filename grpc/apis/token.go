package apis

import (
	"context"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
)

type TokenService struct {
	cfg *config.SignerConfig
}

func NewTokenService(cfg *config.SignerConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

func (t *TokenService) Refresh(ctx context.Context, request *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	if token, err := t.cfg.JwtManager.Refresh(request.Token); err == nil {
		return &pb.RefreshResponse{Token: token}, nil
	} else {
		return nil, err
	}
}
