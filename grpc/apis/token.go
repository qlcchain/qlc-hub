package apis

import (
	"context"
	"fmt"

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

func (t *TokenService) AddressList(ctx context.Context, request *pb.AddressRequest) (*pb.AddressResponse, error) {
	typ := pb.SignType(request.Type)
	list := t.cfg.AddressList(typ)
	if len(list) > 0 {
		return &pb.AddressResponse{Address: list}, nil
	} else {
		return nil, fmt.Errorf("can not find any address of %s", typ.String())
	}
}
