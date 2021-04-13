package apis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	nep5Eth *eth.Transaction
	qgasEth *eth.Transaction
	bsc     *eth.Transaction
	neo     *neo.Transaction
	cfg     *config.Config
	ctx     context.Context
	store   *gorm.DB
	logger  *zap.SugaredLogger
}

func NewInfoAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, nep5Eth, qgasEth, bsc *eth.Transaction, s *gorm.DB) *InfoAPI {
	api := &InfoAPI{
		ctx:     ctx,
		cfg:     cfg,
		neo:     neo,
		nep5Eth: nep5Eth,
		qgasEth: qgasEth,
		bsc:     bsc,
		store:   s,
		logger:  log.NewLogger("api/info"),
	}
	return api
}

func (i *InfoAPI) Ping(ctx context.Context, empty *empty.Empty) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		EthContract: i.cfg.EthCfg.EthNep5Contract,
		EthOwner:    i.cfg.EthCfg.EthNep5Owner,
		EthUrl:      i.nep5Eth.ClientEndpoint(),
		NeoContract: i.cfg.NEOCfg.Contract,
		NeoOwner:    i.cfg.NEOCfg.Owner,
		NeoUrl:      i.neo.ClientEndpoint(),
		QlcOwner:    i.cfg.QlcCfg.QlcOwner,
		TotalSupply: i.nep5Eth.TotalSupply(),
	}, nil
}

func (i *InfoAPI) Config(ctx context.Context, empty *empty.Empty) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{
		NeoContract:     i.cfg.NEOCfg.Contract,
		NeoOwner:        i.cfg.NEOCfg.Owner,
		NeoUrl:          i.neo.ClientEndpoint(),
		EthNep5Contract: i.cfg.EthCfg.EthNep5Contract,
		EthNep5Owner:    i.cfg.EthCfg.EthNep5Owner,
		EthUrl:          i.nep5Eth.ClientEndpoint(),
		EthQGasContract: i.cfg.EthCfg.EthQGasContract,
		EthQGasOwner:    i.cfg.EthCfg.EthQGasOwner,
		BscNep5Contract: i.cfg.BscCfg.BscNep5Contract,
		BscNep5Owner:    i.cfg.BscCfg.BscNep5Owner,
		BscQGasContract: i.cfg.BscCfg.BscQGasContract,
		BscQGasOwner:    i.cfg.BscCfg.BscQGasOwner,
		BscUrl:          i.bsc.ClientEndpoint(),
		QlcOwner:        i.cfg.QlcCfg.QlcOwner,
		QlcUrl:          i.cfg.QlcCfg.EndPoint,
		TotalSupply:     i.nep5Eth.TotalSupply(),
	}, nil
}

func (i *InfoAPI) SwapInfoList(ctx context.Context, offset *pb.Offset) (*pb.SwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetPage(), offset.GetPageSize())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()

	infos, err := db.GetSwapInfos(i.store, offset.GetChain(), int(page), int(pageSize))
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
	chain := offset.GetChain()

	if err := i.neo.ValidateAddress(addr); err == nil {
		infos, err := db.GetSwapInfosByAddr(i.store, int(page), int(pageSize), addr, chain, false)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toSwapInfos(infos), nil
	} else {
		infos, err := db.GetSwapInfosByAddr(i.store, int(page), int(pageSize), addr, chain, true)
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
	infos, err := db.GetSwapInfos(i.store, "", 0, 0)
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
	infos, err := db.GetSwapInfos(i.store, "", 0, 0)
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

func (i *InfoAPI) SwapAmountByAddress(ctx context.Context, address *pb.Address) (*pb.AmountByAddressResponse, error) {
	addr := address.GetAddress()
	if addr == "" {
		return nil, errors.New("invalid params")
	}

	if err := i.neo.ValidateAddress(addr); err == nil {
		infos, err := db.GetSwapInfosByAddr(i.store, 0, 0, addr, "", false)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return i.swapAmountByAddress(infos, addr, false)
	} else {
		infos, err := db.GetSwapInfosByAddr(i.store, 0, 0, addr, "", true)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return i.swapAmountByAddress(infos, addr, true)
	}
}

func (i *InfoAPI) swapAmountByAddress(infos []*types.SwapInfo, addr string, isEthAddr bool) (*pb.AmountByAddressResponse, error) {
	pledgeCount := 0
	var pledgeAmount int64 = 0
	withdrawCount := 0
	var withdrawAmount int64 = 0
	for _, info := range infos {
		if isEthAddr {
			if strings.ToLower(info.EthUserAddr) == strings.ToLower(addr) {
				if info.State == types.DepositDone {
					pledgeCount = pledgeCount + 1
					pledgeAmount = pledgeAmount + info.Amount
				}
				if info.State == types.WithDrawDone {
					withdrawCount = withdrawCount + 1
					withdrawAmount = withdrawAmount + info.Amount
				}
			}
		} else {
			if info.NeoUserAddr == addr {
				if info.State == types.DepositDone {
					pledgeCount = pledgeCount + 1
					pledgeAmount = pledgeAmount + info.Amount
				}
				if info.State == types.WithDrawDone {
					withdrawCount = withdrawCount + 1
					withdrawAmount = withdrawAmount + info.Amount
				}
			}
		}
	}
	var balance int64 = 0
	if isEthAddr {
		b, err := i.nep5Eth.BalanceOf(addr)
		if err == nil && b != nil {
			balance = b.Int64()
		}
	}
	return &pb.AmountByAddressResponse{
		Address:        addr,
		Erc20Amount:    balance,
		PledgeCount:    int64(pledgeCount),
		PledgeAmount:   pledgeAmount,
		WithdrawCount:  int64(withdrawCount),
		WithdrawAmount: withdrawAmount,
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
		NeoUserAddr:    info.NeoUserAddr,
		NeoTxHash:      info.NeoTxHash,
		Chain:          types.ChainTypeToString(info.Chain),
		ChainUserAddr:  info.EthUserAddr,
		ChainTxHash:    info.EthTxHash,
		StartTime:      time.Unix(info.StartTime, 0).Format(time.RFC3339),
		LastModifyTime: time.Unix(info.LastModifyTime, 0).Format(time.RFC3339),
	}
}

func (i *InfoAPI) CheckNeoTransaction(ctx context.Context, Hash *pb.Hash) (*pb.Boolean, error) {
	neoTxHash := Hash.GetHash()
	if neoTxHash == "" {
		return nil, errors.New("invalid params")
	}
	if err := i.neo.TxVerifyAndConfirmed(neoTxHash, i.cfg.NEOCfg.ConfirmedHeight); err != nil {
		return toBoolean(false), err
	}

	//hash, err := util.Uint256DecodeStringLE(hubUtil.RemoveHexPrefix(neoTxHash))
	//if err != nil {
	//	return toBoolean(false), err
	//}
	//neoInfo, err := i.neo.QueryLockedInfo(hash.StringBE())
	//if err != nil || neoInfo == nil {
	//	return toBoolean(false), err
	//}
	return toBoolean(true), nil
}

func (i *InfoAPI) CheckEthTransaction(ctx context.Context, Hash *pb.Hash) (*pb.Boolean, error) {
	hash := common.HexToHash(Hash.GetHash())
	confirmed, err := i.nep5Eth.HasBlockConfirmed(hash, i.cfg.EthCfg.EthConfirmedHeight+1)
	if err != nil || !confirmed {
		return nil, fmt.Errorf("block not confirmed, %s", err)
	}
	if _, _, _, err := i.nep5Eth.SyncBurnLog(Hash.GetHash()); err != nil {
		if _, _, neoTx, err := i.nep5Eth.SyncMintLog(Hash.GetHash()); err != nil {
			return toBoolean(false), fmt.Errorf("no sync log, %s", err)
		} else {
			if _, err := db.GetSwapInfoByTxHash(i.store, neoTx, types.NEO); err != nil {
				return toBoolean(false), err
			} else {
				return toBoolean(true), nil
			}
		}
	} else {
		return toBoolean(true), nil
	}
}

func (i *InfoAPI) SwapUnconfirmeds(ctx context.Context, empty *empty.Empty) (*pb.SwapPendings, error) {
	infos, err := db.GetSwapPendings(i.store, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toSwapPendings(infos), nil
}

func toSwapPendings(infos []*types.SwapPending) *pb.SwapPendings {
	r := make([]*pb.SwapPending, 0)
	for _, info := range infos {
		ri := &pb.SwapPending{
			Typ:            int32(info.Typ),
			EthTxHash:      info.EthTxHash,
			NeoTxHash:      info.NeoTxHash,
			LastModifyTime: time.Unix(info.LastModifyTime, 0).Format(time.RFC3339),
		}
		r = append(r, ri)
	}
	return &pb.SwapPendings{
		Infos: r,
	}
}
