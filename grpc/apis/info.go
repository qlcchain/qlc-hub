package apis

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InfoAPI struct {
	eth    *eth.Transaction
	neo    *neo.Transaction
	cfg    *config.Config
	ctx    context.Context
	store  *gorm.DB
	logger *zap.SugaredLogger
}

func NewInfoAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, eth *eth.Transaction, s *gorm.DB) *InfoAPI {
	return &InfoAPI{
		ctx:    ctx,
		cfg:    cfg,
		neo:    neo,
		eth:    eth,
		store:  s,
		logger: log.NewLogger("api/info"),
	}
}

func (i *InfoAPI) Ping(ctx context.Context, empty *empty.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		EthContract:       i.cfg.EthCfg.Contract,
		EthOwner:          i.cfg.EthCfg.OwnerAddress,
		EthUrl:            i.cfg.EthCfg.EndPoint,
		NeoContract:       i.cfg.NEOCfg.Contract,
		NeoOwner:          i.cfg.NEOCfg.SignerAddress,
		NeoUrl:            i.neo.ClientEndpoint(),
		MinDepositAmount:  i.cfg.MinDepositAmount,
		MinWithdrawAmount: i.cfg.MinWithdrawAmount,
	}, nil
}

func (i *InfoAPI) SwapInfoList(ctx context.Context, offset *pb.Offset) (*pb.SwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetPage(), offset.GetPageSize())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()

	infos, err := db.GetSwapInfos(i.store, int(page), int(pageSize))
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toSwapInfos(infos), nil
}

func (i *InfoAPI) SwapInfosByAddress(ctx context.Context, offset *pb.AddrAndOffset) (*pb.SwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetAddress())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	addr := offset.GetAddress()

	if err := i.neo.ValidateAddress(addr); err == nil {
		infos, err := db.GetSwapInfosByAddr(i.store, int(page), int(pageSize), addr, types.NEO)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toSwapInfos(infos), nil
	} else {
		infos, err := db.GetSwapInfosByAddr(i.store, int(page), int(pageSize), addr, types.ETH)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toSwapInfos(infos), nil
	}
}

func (i *InfoAPI) SwapInfoByTxHash(ctx context.Context, h *pb.Hash) (*pb.SwapInfo, error) {
	hash := h.GetHash()
	if !(len(hash) == 66 || len(hash) == 64) {
		return nil, fmt.Errorf("invalid hash: %s", hash)
	}
	info, err := db.GetSwapInfoByTxHash(i.store, hash, types.ETH)
	if err != nil {
		info, err := db.GetSwapInfoByTxHash(i.store, hash, types.NEO)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		} else {
			return toSwapInfo(info), nil
		}
	} else {
		return toSwapInfo(info), nil
	}
}

func (i *InfoAPI) SwapInfosByState(ctx context.Context, offset *pb.StateAndOffset) (*pb.SwapInfos, error) {
	if types.StringToSwapState(offset.GetState()) == types.Invalid {
		return nil, fmt.Errorf("invalid state: %s", offset.GetState())
	}
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetState())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	state := types.StringToSwapState(offset.GetState())
	infos, err := db.GetSwapInfosByState(i.store, int(page), int(pageSize), state)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toSwapInfos(infos), nil
}

func (i *InfoAPI) SwapCountByState(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	count := make(map[string]int64)
	infos, err := db.GetSwapInfos(i.store, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.DepositDone {
			count["Deposit"] = count["Deposit"] + 1
		} else {
			count["Withdraw"] = count["Withdraw"] + 1
		}
		count[types.SwapStateToString(info.State)] = count[types.SwapStateToString(info.State)] + 1
	}
	return &pb.Map{
		Count: count,
	}, nil
}

func (i *InfoAPI) SwapAmountByState(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	amount := make(map[string]int64)
	infos, err := db.GetSwapInfos(i.store, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.DepositDone {
			amount["Deposit"] = amount["Deposit"] + info.Amount
		} else {
			amount["Withdraw"] = amount["Withdraw"] + info.Amount
		}
		amount[types.SwapStateToString(info.State)] = amount[types.SwapStateToString(info.State)] + info.Amount
	}
	return &pb.Map{
		Count: amount,
	}, nil
}

func toSwapInfos(infos []*types.SwapInfo) *pb.SwapInfos {
	r := make([]*pb.SwapInfo, 0)
	for _, info := range infos {
		r = append(r, toSwapInfo(info))
	}
	return &pb.SwapInfos{
		Infos: r,
	}
}

func toSwapInfo(info *types.SwapInfo) *pb.SwapInfo {
	return &pb.SwapInfo{
		State:          int32(info.State),
		StateStr:       types.SwapStateToString(info.State),
		Amount:         info.Amount,
		EthTxHash:      info.EthTxHash,
		NeoTxHash:      info.NeoTxHash,
		EthUserAddr:    info.EthUserAddr,
		NeoUserAddr:    info.NeoUserAddr,
		StartTime:      time.Unix(info.StartTime, 0).Format(time.RFC3339),
		LastModifyTime: time.Unix(info.LastModifyTime, 0).Format(time.RFC3339),
	}
}
