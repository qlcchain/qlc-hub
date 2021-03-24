package apis

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	qlcchain "github.com/qlcchain/qlc-go-sdk"
	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/config"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/qlc"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type QGasSwapAPI struct {
	eth       *eth.Transaction
	qlc       *qlc.Transaction
	store     *gorm.DB
	cfg       *config.Config
	ctx       context.Context
	ownerAddr qlctypes.Address
	addrPool  *AddressPool
	signer    *signer.SignerClient
	logger    *zap.SugaredLogger
}

func NewQGasSwapAPI(ctx context.Context, cfg *config.Config, q *qlc.Transaction, e *eth.Transaction, signer *signer.SignerClient, s *gorm.DB) *QGasSwapAPI {
	api := &QGasSwapAPI{
		cfg:    cfg,
		eth:    e,
		qlc:    q,
		ctx:    ctx,
		store:  s,
		signer: signer,
		logger: log.NewLogger("api/qgas_pledge"),
	}
	address, err := qlctypes.HexToAddress(cfg.QlcCfg.OwnerAddress)
	if err != nil {
		api.logger.Fatal(err)
	}
	api.ownerAddr = address
	api.addrPool = AddressPools(address)
	return api
}

type QGasPledgeParam struct {
	PledgeAddress     qlctypes.Address
	Amount            qlctypes.Balance
	Erc20ReceiverAddr string
}

func (g *QGasSwapAPI) GetPledgeBlock(ctx context.Context, params *pb.QGasPledgeRequest) (*pb.StateBlockHash, error) {
	g.logger.Infof("call QGas Get Pledge Send Block......... (%s) ", params)
	if params.GetPledgeAddress() == "" || params.Erc20ReceiverAddr == "" || params.GetAmount() <= 0 {
		return nil, errors.New("error params")
	}

	pledgeAddress, err := qlctypes.HexToAddress(params.GetPledgeAddress())
	if err != nil {
		g.logger.Errorf("invalid address, %s, %s", err, params.GetPledgeAddress())
		return nil, err
	}

	sendBlk, err := g.qlc.Client().QGasSwap.GetPledgeSendBlock(&qlcchain.QGasPledgeParam{
		PledgeAddress: pledgeAddress,
		Amount:        qlctypes.Balance{Int: big.NewInt(params.GetAmount())},
		ToAddress:     g.ownerAddr,
	})
	if err != nil {
		g.logger.Errorf("QGas get pledge send block: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas create pledge send block: %s", sendBlk.GetHash())

	swapInfo := &types.QGasSwapInfo{
		SwapType:     types.QGasDeposit,
		State:        types.QGasPledgeInit,
		Amount:       params.GetAmount(),
		QlcUserAddr:  pledgeAddress.String(),
		OwnerAddress: g.ownerAddr.String(),
		SendTxHash:   sendBlk.GetHash().String(),
		BlockStr:     toBlockStr(sendBlk),
		UserTxHash:   sendBlk.GetHash().String(),
		EthUserAddr:  params.GetErc20ReceiverAddr(),
		StartTime:    time.Now().Unix(),
	}

	if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
		g.logger.Errorf("insert invalid info: %s", err)
		return nil, err
	}
	g.logger.Infof("QGas insert pledge info to %s, qlc[%s]", types.QGasSwapStateToString(types.QGasPledgeInit), sendBlk.GetHash())

	return &pb.StateBlockHash{
		Hash: sendBlk.GetHash().String(),
		Root: sendBlk.Root().String(),
	}, nil
}

func toBlockStr(blk *qlctypes.StateBlock) string {
	blkBytes, _ := blk.Serialize()
	return hex.EncodeToString(blkBytes)
}

func (g *QGasSwapAPI) GetWithdrawBlock(ctx context.Context, param *pb.Hash) (*pb.StateBlockHash, error) {
	g.logger.Infof("call QGas Get Withdraw Block ......... (%s) ", param)
	if param == nil {
		return nil, errors.New("error params")
	}
	ethTxHash := param.GetHash()
	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, ethTxHash, types.ETH)
	if err != nil {
		g.logger.Errorf("QGas eth tx not found", err)
		return nil, err
	} else {
		if swapInfo.State < types.QGasWithDrawPending {
			g.logger.Errorf("invalid state: %s, eth[%s]", types.QGasSwapStateToString(swapInfo.State), ethTxHash)
			return nil, errors.New("invalid  state")
		}
		if swapInfo.State == types.QGasWithDrawDone {
			g.logger.Errorf("reduplicate tx: %s, eth[%s]", err, ethTxHash)
			return nil, errors.New("reduplicate tx")
		}
	}

	if swapInfo.SendTxHash == "" {
		qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
		if qlcAddress == qlctypes.ZeroAddress {
			g.logger.Errorf("can not search address %s, eth[%s]", g.ownerAddr, ethTxHash)
			return nil, errors.New("can not search address")
		}
		defer g.addrPool.Enqueue(qlcAddress)

		qlcUserAddress, err := qlctypes.HexToAddress(swapInfo.QlcUserAddr)
		if err != nil {
			g.logger.Errorf("QGas invalid address: %s, %s, eth[%s]", err, swapInfo.QlcUserAddr, ethTxHash)
			return nil, err
		}
		var linkHash qlctypes.Hash
		if err := linkHash.Of(hubUtil.RemoveHexPrefix(ethTxHash)); err != nil {
			g.logger.Errorf("QGas invalid eth tx: %s,  eth[%s]", err, ethTxHash)
			return nil, err
		}
		sendBlk, err := g.qlc.Client().QGasSwap.GetWithdrawSendBlock(&qlcchain.QGasWithdrawParam{
			WithdrawAddress: qlcUserAddress,
			Amount:          qlctypes.Balance{Int: big.NewInt(swapInfo.Amount)},
			FromAddress:     g.ownerAddr,
			LinkHash:        linkHash,
		})
		if err != nil {
			g.logger.Errorf("QGas get withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		if err := g.signQLCTx(sendBlk); err != nil {
			g.logger.Errorf("QGas sign reward block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		if err := g.qlc.ProcessAndWaitConfirmed(sendBlk); err != nil {
			g.logger.Errorf("QGas process withdraw send block: %s, eth[%s]", err, ethTxHash)
			return nil, err
		}

		swapInfo.SendTxHash = sendBlk.GetHash().String()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("update invalid info: %s", err)
			return nil, err
		}
		g.logger.Infof("QGas withdraw send block successfully, %s, eth[%s]", sendBlk.GetHash(), ethTxHash)
	}

	sendTxHash, err := stringToHash(swapInfo.SendTxHash)
	if err != nil {
		g.logger.Errorf("QGas invalid withdraw send hash: %s, %s, eth[%s]", err, swapInfo.SendTxHash, ethTxHash)
		return nil, err
	}
	rewardBlk, err := g.qlc.Client().QGasSwap.GetWithdrawRewardBlock(sendTxHash)
	if err != nil {
		g.logger.Errorf("QGas get withdraw reward block: %s, eth[%s]", err, ethTxHash)
		return nil, err
	}

	swapInfo.BlockStr = toBlockStr(rewardBlk)
	swapInfo.UserTxHash = rewardBlk.GetHash().String()
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		return nil, err
	}

	return &pb.StateBlockHash{
		Hash: rewardBlk.GetHash().String(),
		Root: rewardBlk.Root().String(),
	}, nil
}

func (g *QGasSwapAPI) convertBlock(params *pb.StateBlockSigned) (*qlctypes.StateBlock, *types.QGasSwapInfo, error) {
	signStr := params.GetSignature()
	signature, err := qlctypes.NewSignature(signStr)
	if err != nil {
		return nil, nil, fmt.Errorf("signature, %s", err)
	}
	workStr := params.GetWork()
	work := new(qlctypes.Work)
	if err := work.ParseWorkHexString(workStr); err != nil {
		return nil, nil, fmt.Errorf("work, %s", err)
	}

	swapInfo, err := db.GetQGasSwapInfoByUserTxHash(g.store, params.GetHash())
	blockStr := swapInfo.BlockStr
	blockBytes, err := hex.DecodeString(blockStr)
	if err != nil {
		return nil, nil, fmt.Errorf("decode string, %s", err)
	}
	block := new(qlctypes.StateBlock)
	if err := block.Deserialize(blockBytes); err != nil {
		return nil, nil, fmt.Errorf("block deserialize, %s", err)
	}
	block.Work = *work
	block.Signature = signature
	return block, swapInfo, nil
}

func (g *QGasSwapAPI) ProcessBlock(ctx context.Context, params *pb.StateBlockSigned) (*pb.Hash, error) {
	if params == nil {
		return nil, errors.New("nil block")
	}
	g.logger.Infof("call QGas Process Block ......... (%s) ", params)
	if params.GetHash() == "" || params.GetSignature() == "" || params.GetWork() == "" {
		return nil, errors.New("invalid params")
	}

	blk, swapInfo, err := g.convertBlock(params)
	if err != nil {
		g.logger.Errorf("QGas get block: %s, %s", err, params.GetHash())
		return nil, err
	}

	if blk.Type == qlctypes.ContractSend { // pledge
		if blk.GetHash().String() != swapInfo.SendTxHash {
			g.logger.Errorf("QGas invalid send hash: %s, qlc[%s]", swapInfo.SendTxHash, blk.GetHash())
			return nil, errors.New("invalid state")
		}

		if swapInfo.State >= types.QGasPledgeDone {
			g.logger.Errorf("QGas invalid pledge state: %s, qlc[%s]", types.QGasSwapStateToString(swapInfo.State), blk.GetHash())
			return nil, errors.New("invalid state")
		}

		if swapInfo.State == types.QGasPledgeInit {
			if !g.qlc.CheckBlockOnChain(blk.GetHash()) {
				if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
					g.logger.Errorf("QGas Process pledge send block: %s [%s]", err, blk.GetHash())
					return nil, err
				}
				g.logger.Infof("QGas process pledge send block successfully, qlc[%s]", blk.GetHash())
			}

			qlcAddress := g.addrPool.SearchSync(g.ownerAddr)
			if qlcAddress == qlctypes.ZeroAddress {
				g.logger.Errorf("can not search address %s, qlc[%s]", g.ownerAddr, blk.GetHash())
				return nil, errors.New("can not search address")
			}
			defer g.addrPool.Enqueue(qlcAddress)

			rewardBlk, err := g.qlc.Client().QGasSwap.GetPledgeRewardBlock(blk.GetHash())
			if err != nil {
				g.logger.Errorf("QGas get pledge reward block error: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}

			if err := g.signQLCTx(rewardBlk); err != nil {
				g.logger.Errorf("QGas sign reward block: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}

			if err := g.qlc.ProcessAndWaitConfirmed(rewardBlk); err != nil {
				g.logger.Errorf("QGas Process pledge reward block: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}
			g.logger.Infof("QGas process pledge reward block successfully: %s, qlc[%s]", rewardBlk.GetHash().String(), blk.GetHash())

			swapInfo.RewardTxHash = rewardBlk.GetHash().String()
			swapInfo.State = types.QGasPledgePending
			if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
				g.logger.Errorf("update invalid info: %s, qlc[%s]", err, blk.GetHash())
				return nil, err
			}
			g.logger.Infof("QGas update pledge info to %s: qlc[%s]", types.QGasSwapStateToString(types.QGasPledgePending), swapInfo.SendTxHash)
			return &pb.Hash{
				Hash: rewardBlk.GetHash().String(),
			}, nil
		}

		if swapInfo.State == types.QGasPledgePending {
			return &pb.Hash{
				Hash: swapInfo.RewardTxHash,
			}, nil
		}
	} else if blk.Type == qlctypes.ContractReward { // withdraw
		if !g.qlc.CheckBlockOnChain(blk.GetHash()) {
			if err := g.qlc.ProcessAndWaitConfirmed(blk); err != nil {
				g.logger.Errorf("QGas Process withdraw reward block: %s, eth[%s]", err, swapInfo.EthTxHash)
				return nil, err
			}
			g.logger.Infof("QGas process pledge reward block successfully: %s, eth[%s]", blk.GetHash().String(), swapInfo.EthTxHash)

			swapInfo.RewardTxHash = blk.GetHash().String()
			swapInfo.State = types.QGasWithDrawDone
			if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
				g.logger.Errorf("update invalid info: %s", err)
				return nil, err
			}
			g.logger.Infof("QGas update withdraw info to %s: eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawDone), swapInfo.EthTxHash)
		}
		return &pb.Hash{
			Hash: swapInfo.RewardTxHash,
		}, nil
	}
	return nil, fmt.Errorf("invalid block typ: %s", blk.GetType())
}

func (g *QGasSwapAPI) GetEthOwnerSign(ctx context.Context, param *pb.Hash) (*pb.String, error) {
	txHash := param.GetHash()
	if txHash == "" {
		g.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}
	g.logger.Infof("call QGas GetEthOwnerSign: %s", param)

	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, txHash, types.QLC)
	if err != nil {
		g.logger.Infof("qlc not locked, %s, qlc[%s]", err, txHash)
		return nil, fmt.Errorf("qlc not locked")
	}
	if swapInfo.State >= types.QGasPledgeDone {
		g.logger.Errorf("repeat operation, qlc[%s]", txHash)
		return nil, fmt.Errorf("repeat operation, qlc[%s]", txHash)
	}

	sign, err := g.signEthData(big.NewInt(swapInfo.Amount), swapInfo.EthUserAddr, hubUtil.RemoveHexPrefix(txHash))
	if err != nil {
		g.logger.Error(err)
		return nil, err
	}
	g.logger.Infof("QGas hub signed: %s. qlc[%s]", sign, txHash)
	return toString(sign), nil
}

func (g *QGasSwapAPI) PledgeEthTxSent(ctx context.Context, param *pb.EthTxSentRequest) (*pb.Boolean, error) {
	if param == nil {
		return nil, errors.New("nil param")
	}
	g.logger.Infof("call QGas pledge EthTransactionSent: %s", param)
	qlcTxHash := param.GetQlcTxHash()
	ethTxHash := param.GetEthTxHash()

	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, qlcTxHash, types.QLC)
	if err != nil {
		g.logger.Errorf("get swap info error: %s, qlc[%s]", err, qlcTxHash)
		return nil, errors.New("swap info not found")
	}
	if swapInfo.State >= types.QGasPledgeDone {
		return toBoolean(true), nil
	}
	if swapInfo.State < types.QGasPledgePending {
		g.logger.Errorf("invalid pledge state: %s", types.QGasSwapStateToString(swapInfo.State))
		return nil, errors.New("invalid state")
	}
	go func() {
		if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(ethTxHash), 0, g.cfg.EthCfg.ConfirmedHeight); err != nil {
			g.logger.Errorf("QGas pledge eth tx confirmed: %s", err)
			return
		}
		g.logger.Infof("QGas pledge eth Tx confirmed, %s, qlc[%s]", ethTxHash, qlcTxHash)

		amount, ethAddress, qlcTx, err := g.eth.QGasSyncMintLog(ethTxHash)
		if err != nil {
			g.logger.Errorf("mint log: %s", err)
			return
		}

		if amount.Int64() != swapInfo.Amount || strings.ToLower(ethAddress.String()) != strings.ToLower(swapInfo.EthUserAddr) || qlcTx != qlcTxHash {
			g.logger.Errorf("swap info not match: %s, amount %d, address %s", qlcTx, amount.Int64(), ethAddress.String())
			return
		}

		swapInfo.State = types.QGasPledgeDone
		swapInfo.EthTxHash = ethTxHash
		swapInfo.EthUserAddr = ethAddress.String()
		swapInfo.Amount = amount.Int64()
		if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Error(err)
			return
		}
		g.logger.Infof("update state to %s, qlc[%s]", types.QGasSwapStateToString(types.QGasPledgeDone), qlcTx)
		g.logger.Infof("QGas pledge successfully. qlc[%s]", qlcTx)
	}()
	return toBoolean(true), nil
}

func (g *QGasSwapAPI) WithdrawEthTxSent(ctx context.Context, param *pb.Hash) (*pb.Boolean, error) {
	if param == nil {
		return nil, errors.New("nil param")
	}
	g.logger.Infof("call QGas withdraw EthTransactionSent: %s", param.GetHash())
	ethTxHash := param.GetHash()

	_, err := db.GetQGasSwapInfoByUniqueID(g.store, ethTxHash, types.ETH)
	if err == nil {
		return toBoolean(true), nil
	}

	go func() {
		swapInfo := &types.QGasSwapInfo{
			SwapType:     types.QGasWithdraw,
			State:        types.QGasWithDrawInit,
			EthTxHash:    ethTxHash,
			OwnerAddress: g.ownerAddr.String(),
			StartTime:    time.Now().Unix(),
		}
		if err := db.InsertQGasSwapInfo(g.store, swapInfo); err != nil {
			g.logger.Errorf("QGas insert invalid info: %s, eth[%s]", err, ethTxHash)
			return
		}
		g.logger.Infof("QGas insert withdraw info to %s, eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawInit), ethTxHash)

		if err := g.withdrawInit(ethTxHash); err != nil {
			g.logger.Errorf("QGas eth tx confirmed:  %s, eth[%s]", err, ethTxHash)
		}
	}()
	return toBoolean(true), nil
}

func (g *QGasSwapAPI) withdrawInit(ethTxHash string) error {
	if err := g.eth.WaitTxVerifyAndConfirmed(common.HexToHash(ethTxHash), 0, g.cfg.EthCfg.ConfirmedHeight); err != nil {
		return fmt.Errorf("tx confirmed: %s", err)
	}
	g.logger.Infof("QGas withdraw eth tx confirmed, eth[%s]", ethTxHash)
	amount, user, qlcAddrStr, err := g.eth.QGasSyncBurnLog(ethTxHash)
	if err != nil {
		return fmt.Errorf("get burn log, %s", err)
	}

	qlcAddr, err := qlctypes.HexToAddress(qlcAddrStr)
	if err != nil {
		return err
	}

	swapInfo, err := db.GetQGasSwapInfoByUniqueID(g.store, ethTxHash, types.ETH)
	if err != nil {
		return fmt.Errorf("get swap info, %s", err)
	}

	swapInfo.Amount = amount.Int64()
	swapInfo.QlcUserAddr = qlcAddr.String()
	swapInfo.EthUserAddr = user.String()
	swapInfo.State = types.QGasWithDrawPending
	if err := db.UpdateQGasSwapInfo(g.store, swapInfo); err != nil {
		return err
	}
	g.logger.Infof("QGas update withdraw info to %s, eth[%s]", types.QGasSwapStateToString(types.QGasWithDrawPending), ethTxHash)
	return nil
}

func (g *QGasSwapAPI) signEthData(amount *big.Int, receiveAddr string, neoTxHash string) (string, error) {
	packedHash, err := packed(amount, receiveAddr, neoTxHash)
	if err != nil {
		return "", fmt.Errorf("packed: %s", err)
	}

	signature, err := g.signer.Sign(pb.SignType_ETH, g.cfg.EthCfg.OwnerAddress, packedHash)
	if err != nil {
		return "", fmt.Errorf("sign: %s", err)
	}
	sig := signature.Sign
	if len(sig) == 0 {
		return "", errors.New("invalid signature")
	}

	v := sig[len(sig)-1]
	if v == 0 || v == 1 {
		sig[len(sig)-1] = v + 27
		return hex.EncodeToString(sig), nil
	} else if v == 27 || v == 28 {
		return hex.EncodeToString(sig), nil
	} else {
		return "", fmt.Errorf("invalid signature 'v' value: %s", hex.EncodeToString(sig))
	}
}

func (g *QGasSwapAPI) signQLCTx(block *qlctypes.StateBlock) error {
	var w qlctypes.Work
	worker, err := qlctypes.NewWorker(w, block.Root())
	if err != nil {
		return err
	}
	block.Work = worker.NewWork()
	hash := block.GetHash()
	signature, err := g.signer.Sign(pb.SignType_QLC, g.cfg.QlcCfg.OwnerAddress, hash.Bytes())
	if err != nil {
		return fmt.Errorf("sign: %s", err)
	}
	sign, err := qlctypes.BytesToSignature(signature.GetSign())
	if err != nil {
		return fmt.Errorf("sign bytes: %s", err)
	}
	block.Signature = sign
	return nil
}

func stringToHash(h string) (qlctypes.Hash, error) {
	var hash qlctypes.Hash
	if err := hash.Of(h); err != nil {
		return qlctypes.ZeroHash, err
	}
	return hash, nil
}

func toQGasSwapInfo(info *types.QGasSwapInfo) *pb.QGasSwapInfo {
	return &pb.QGasSwapInfo{
		State:          int32(info.State),
		StateStr:       types.QGasSwapStateToString(info.State),
		Amount:         info.Amount,
		EthTxHash:      info.EthTxHash,
		SendTxHash:     info.SendTxHash,
		RewardTxHash:   info.RewardTxHash,
		EthUserAddr:    info.EthUserAddr,
		QlcUserAddr:    info.QlcUserAddr,
		StartTime:      time.Unix(info.StartTime, 0).Format(time.RFC3339),
		LastModifyTime: time.Unix(info.LastModifyTime, 0).Format(time.RFC3339),
	}
}

func toQGasSwapInfos(infos []*types.QGasSwapInfo) *pb.QGasSwapInfos {
	r := make([]*pb.QGasSwapInfo, 0)
	for _, info := range infos {
		r = append(r, toQGasSwapInfo(info))
	}
	return &pb.QGasSwapInfos{
		Infos: r,
	}
}

func (g *QGasSwapAPI) SwapInfoByTxHash(ctx context.Context, h *pb.Hash) (*pb.QGasSwapInfo, error) {
	hash := h.GetHash()
	if !(len(hash) == 66 || len(hash) == 64) {
		return nil, fmt.Errorf("invalid hash: %s", hash)
	}
	info, err := db.GetQGasSwapInfoByUniqueID(g.store, hash, types.ETH)
	if err != nil {
		info, err := db.GetQGasSwapInfoByUniqueID(g.store, hash, types.QLC)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		} else {
			return toQGasSwapInfo(info), nil
		}
	} else {
		return toQGasSwapInfo(info), nil
	}
}

func (g *QGasSwapAPI) SwapInfoList(ctx context.Context, offset *pb.Offset) (*pb.QGasSwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d", offset.GetPage(), offset.GetPageSize())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()

	infos, err := db.GetQGasSwapInfos(g.store, int(page), int(pageSize))
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toQGasSwapInfos(infos), nil
}

func (g *QGasSwapAPI) SwapInfosByAddress(ctx context.Context, offset *pb.AddrAndOffset) (*pb.QGasSwapInfos, error) {
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetAddress())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	addr := offset.GetAddress()

	if err := g.qlc.ValidateAddress(addr); err == nil {
		infos, err := db.GetQGasSwapInfosByAddr(g.store, int(page), int(pageSize), addr, types.QLC)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toQGasSwapInfos(infos), nil
	} else {
		infos, err := db.GetQGasSwapInfosByAddr(g.store, int(page), int(pageSize), addr, types.ETH)
		if err != nil {
			return nil, fmt.Errorf("get swapInfos: %s", err)
		}
		return toQGasSwapInfos(infos), nil
	}
}

func (g *QGasSwapAPI) SwapInfosByState(ctx context.Context, offset *pb.StateAndOffset) (*pb.QGasSwapInfos, error) {
	if types.StringToQGasSwapState(offset.GetState()) == types.QGasInvalid {
		return nil, fmt.Errorf("invalid state: %s", offset.GetState())
	}
	if offset.GetPage() < 0 || offset.GetPageSize() < 0 {
		return nil, fmt.Errorf("invalid offset, %d, %d, %s", offset.GetPage(), offset.GetPageSize(), offset.GetState())
	}
	page := offset.GetPage()
	pageSize := offset.GetPageSize()
	state := types.StringToQGasSwapState(offset.GetState())
	infos, err := db.GetQGasSwapInfosByState(g.store, int(page), int(pageSize), state)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	return toQGasSwapInfos(infos), nil
}

func (g *QGasSwapAPI) SwapInfosCount(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	count := make(map[string]int64)
	infos, err := db.GetQGasSwapInfos(g.store, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.QGasPledgeDone {
			count["QGasPledgeTotal"] = count["QGasPledgeTotal"] + 1
		} else {
			count["QGasWithdrawTotal"] = count["QGasWithdrawTotal"] + 1
		}
		count[types.QGasSwapStateToString(info.State)] = count[types.QGasSwapStateToString(info.State)] + 1
	}
	return &pb.Map{
		Count: count,
	}, nil
}

func (g *QGasSwapAPI) SwapInfosAmount(ctx context.Context, empty *empty.Empty) (*pb.Map, error) {
	amount := make(map[string]int64)
	infos, err := db.GetQGasSwapInfos(g.store, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("get swapInfos: %s", err)
	}
	for _, info := range infos {
		if info.State <= types.QGasPledgeDone {
			amount["QGasPledgeTotal"] = amount["QGasPledgeTotal"] + info.Amount
		} else {
			amount["QGasWithdrawTotal"] = amount["QGasWithdrawTotal"] + info.Amount
		}
		amount[types.QGasSwapStateToString(info.State)] = amount[types.QGasSwapStateToString(info.State)] + info.Amount
	}
	return &pb.Map{
		Count: amount,
	}, nil
}
