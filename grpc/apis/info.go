package apis

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/store"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

type InfoAPI struct {
	cfg          *config.Config
	ctx          context.Context
	nep5Account  *wallet.Account
	erc20Account common.Address
	store        *store.Store
	logger       *zap.SugaredLogger
}

func NewInfoAPI(ctx context.Context, cfg *config.Config) (*InfoAPI, error) {
	nep5Account, err := wallet.NewAccountFromWIF(cfg.NEOCfg.WIF)
	if err != nil {
		return nil, fmt.Errorf("NewDebugAPI/NewAccountFromWIF: %s", err)
	}
	_, address, err := eth.GetAccountByPriKey(cfg.EthereumCfg.Account)
	if err != nil {
		return nil, fmt.Errorf("NewDebugAPI/GetAccountByPriKey: %s", err)
	}
	store, err := store.NewStore(cfg.DataDir())
	if err != nil {
		return nil, err
	}
	return &InfoAPI{
		ctx:          ctx,
		cfg:          cfg,
		nep5Account:  nep5Account,
		erc20Account: address,
		store:        store,
		logger:       log.NewLogger("api/debug"),
	}, nil
}

func (i *InfoAPI) Ping(ctx context.Context, e *empty.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		NeoContract: i.cfg.NEOCfg.Contract,
		NeoAddress:  i.nep5Account.Address,
		EthContract: i.cfg.EthereumCfg.Contract,
		EthAddress:  i.erc20Account.String(),
	}, nil
}

func (i *InfoAPI) LockerState(ctx context.Context, s *pb.String) (*pb.LockerStateResponse, error) {
	r, err := i.store.GetLockerInfo(s.GetValue())
	if err != nil {
		return nil, err
	}
	return toLockerState(r), nil
}

func (i *InfoAPI) LockerStates(ctx context.Context, offset *pb.Offset) (*pb.LockerStatesResponse, error) {
	panic("implement me")
}

func toLockerState(s *types.LockerInfo) *pb.LockerStateResponse {
	return &pb.LockerStateResponse{
		State:               int64(s.State),
		StateStr:            types.LockerStateToString(s.State),
		RHash:               s.RHash,
		ROrigin:             s.ROrigin,
		Amount:              s.Amount,
		UserErc20Addr:       s.Erc20Addr,
		UserNep5Addr:        s.Nep5Addr,
		LockedNep5Hash:      s.LockedNep5Hash,
		LockedNep5Height:    s.LockedNep5Height,
		LockedErc20Hash:     s.LockedErc20Hash,
		LockedErc20Height:   s.LockedErc20Height,
		UnlockedNep5Hash:    s.UnlockedNep5Hash,
		UnlockedNep5Height:  s.UnlockedNep5Height,
		UnlockedErc20Hash:   s.UnlockedErc20Hash,
		UnlockedErc20Height: s.UnlockedErc20Height,
		Fail:                s.Fail,
		Remark:              s.Remark,
	}
}
