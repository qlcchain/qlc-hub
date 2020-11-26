package apis

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
)

type SignerService struct {
	cfg *config.SignerConfig
}

func NewSignerService(cfg *config.SignerConfig) *SignerService {
	return &SignerService{
		cfg: cfg,
	}
}

func (s *SignerService) Sign(ctx context.Context, request *pb.SignRequest) (*pb.SignResponse, error) {
	address := request.Address
	rawData := request.RawData
	t := request.Type
	if address == "" {
		return nil, errors.New("invalid address")
	}

	if len(rawData) == 0 {
		return nil, errors.New("invalid rawData")
	}
	cache := s.cfg.Keys
	if v, ok := cache[t]; ok {
		if key, ok := v[address]; ok {
			switch k := key.(type) {
			case *ecdsa.PrivateKey:
				if signature, err := crypto.Sign(rawData, k); err == nil {
					return &pb.SignResponse{Sign: signature}, nil
				} else {
					return nil, err
				}
			case *keys.PrivateKey:
				sign := k.Sign(rawData)
				return &pb.SignResponse{
					Sign:       sign,
					VerifyData: k.PublicKey().GetVerificationScript(),
				}, nil
			default:
				return nil, invalidKey(t, address)
			}
		} else {
			return nil, invalidKey(t, address)
		}
	} else {
		return nil, invalidKey(t, address)
	}
}

func invalidKey(t pb.SignType, address string) error {
	return fmt.Errorf("can not find any private key for [%s]%s", t.String(), address)
}
