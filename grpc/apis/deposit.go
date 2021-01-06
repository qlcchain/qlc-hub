package apis

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DepositAPI struct {
	neo     *neo.Transaction
	eth     *eth.Transaction
	store   *gorm.DB
	cfg     *config.Config
	ctx     context.Context
	Account *ecdsa.PrivateKey
	signer  *signer.SignerClient
	logger  *zap.SugaredLogger
}

func NewDepositAPI(ctx context.Context, cfg *config.Config, neo *neo.Transaction, e *eth.Transaction, signer *signer.SignerClient, s *gorm.DB) *DepositAPI {
	api := &DepositAPI{
		cfg:    cfg,
		neo:    neo,
		eth:    e,
		ctx:    ctx,
		store:  s,
		signer: signer,
		logger: log.NewLogger("api/deposit"),
	}
	go api.correctSwapPending()
	return api
}

func (d *DepositAPI) PackNeoTransaction(ctx context.Context, request *pb.PackNeoTxRequest) (*pb.PackNeoTxResponse, error) {
	d.logger.Infof("call deposit PackNeoTransaction: %s", request.String())
	receiverAddr := request.GetErc20ReceiverAddr()
	senderAddr := request.GetNep5SenderAddr()
	amount := request.GetAmount()
	if receiverAddr == "" || senderAddr == "" || amount <= 0 {
		d.logger.Error("unsigned invalid params")
		return nil, errors.New("invalid params")
	}

	txHash, data, err := d.neo.UnsignedLockTransaction(senderAddr, receiverAddr, int(amount))
	if err != nil {
		d.logger.Errorf("unsigned tx: %s", err)
		return nil, err
	}
	d.logger.Infof("pack unsigned tx, data:%s, neo[%s]", data, txHash)
	return &pb.PackNeoTxResponse{
		TxHash:       txHash,
		UnsignedData: data,
	}, nil
}

func (d *DepositAPI) SendNeoTransaction(ctx context.Context, request *pb.SendNeoTxnRequest) (*pb.Boolean, error) {
	d.logger.Infof("call deposit SendNeoTransaction: %s", request.String())
	neoTxHash := request.GetTxHash()
	signature := request.GetSignature()
	publicKey := request.GetPublicKey()
	address := request.GetNep5SenderAddr()
	if neoTxHash == "" || signature == "" || publicKey == "" || address == "" {
		d.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}

	if _, err := db.GetSwapInfoByTxHash(d.store, neoTxHash, types.NEO); err == nil {
		d.logger.Errorf("deposit repeatedly, neo tx[%s]", neoTxHash)
		return nil, fmt.Errorf("deposit repeatedly, tx[%s]", neoTxHash)
	}

	tx, err := d.neo.SendLockTransaction(neoTxHash, signature, publicKey, address)
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}
	if tx != neoTxHash {
		d.logger.Errorf("neo tx hash mismatch,%s, %s", tx, neoTxHash)
		return nil, errors.New("neo tx hash mismatch")
	}
	d.logger.Infof("send neo transaction successfully. neo[%s]", tx)

	go func() {
		if err := d.neoTransactionConfirmed(neoTxHash); err != nil {
			d.logger.Errorf("%s, neo[%s]", err, neoTxHash)
			return
		}
		d.neo.SwapEnd(neoTxHash)
	}()
	return &pb.Boolean{
		Value: true,
	}, nil
}

func (d *DepositAPI) neoTransactionConfirmed(neoTxHash string) error {
	_, err := d.neo.WaitTxVerifyAndConfirmed(neoTxHash, d.cfg.NEOCfg.ConfirmedHeight)
	if err != nil {
		return err
	}
	d.logger.Infof("neo transaction confirmed. neo[%s]", neoTxHash)

	hash, err := util.Uint256DecodeStringLE(hubUtil.RemoveHexPrefix(neoTxHash))
	if err != nil {
		return fmt.Errorf("decode hash: %s", err)
	}
	neoInfo, err := d.neo.QueryLockedInfo(hash.StringBE())
	if err != nil || neoInfo == nil {
		return err
	}
	d.logger.Infof("get locked info: %s. neo[%s]", neoInfo.String(), neoTxHash)

	swapInfo := &types.SwapInfo{
		State:       types.DepositPending,
		Amount:      neoInfo.Amount,
		EthTxHash:   "",
		NeoTxHash:   neoTxHash,
		EthUserAddr: neoInfo.UserEthAddress,
		NeoUserAddr: neoInfo.FromAddress,
		StartTime:   time.Now().Unix(),
	}
	d.logger.Infof("add state to %s, neo[%s]", types.SwapStateToString(types.DepositPending), neoTxHash)
	return db.InsertSwapInfo(d.store, swapInfo)
}

func (d *DepositAPI) NeoTransactionConfirmed(ctx context.Context, request *pb.Hash) (*pb.Boolean, error) {
	d.logger.Infof("call deposit NeoTransactionConfirmed: %s", request.String())
	neoTxHash := request.GetHash()
	if neoTxHash == "" {
		d.logger.Errorf("transaction invalid params")
		return nil, errors.New("invalid params")
	}

	if _, err := db.GetSwapInfoByTxHash(d.store, neoTxHash, types.NEO); err == nil {
		d.logger.Errorf("deposit repeatedly, neo tx[%s]", neoTxHash)
		return nil, fmt.Errorf("deposit repeatedly, tx[%s]", neoTxHash)
	}

	go func() {
		if err := d.neoTransactionConfirmed(neoTxHash); err != nil {
			d.logger.Errorf("%s, neo[%s]", err, neoTxHash)
			return
		}
	}()

	return &pb.Boolean{
		Value: true,
	}, nil
}

func (d *DepositAPI) GetEthOwnerSign(ctx context.Context, request *proto.Hash) (*proto.String, error) {
	d.logger.Infof("call deposit GetEthOwnerSign: %s", request.String())
	neoTxHash := request.GetHash()
	if neoTxHash == "" {
		d.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}

	swapInfo, err := db.GetSwapInfoByTxHash(d.store, neoTxHash, types.NEO)
	if err != nil {
		d.logger.Infof("neo not locked, %s, neo[%s]", err, neoTxHash)
		return nil, fmt.Errorf("neo not locked")
	}
	if swapInfo.State >= types.DepositDone {
		d.logger.Errorf("repeat operation, neo tx[%s]", neoTxHash)
		return nil, fmt.Errorf("repeat operation, [%s]", neoTxHash)
	}

	sign, err := d.signData(big.NewInt(swapInfo.Amount), swapInfo.EthUserAddr, hubUtil.RemoveHexPrefix(neoTxHash))
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}
	d.logger.Infof("hub signed: %s. neo[%s]", sign, neoTxHash)
	return toString(sign), nil
}

func (d *DepositAPI) signData(amount *big.Int, receiveAddr string, neoTxHash string) (string, error) {
	packedHash, err := packed(amount, receiveAddr, neoTxHash)
	if err != nil {
		return "", fmt.Errorf("packed: %s", err)
	}

	signature, err := d.signer.Sign(proto.SignType_ETH, d.cfg.EthCfg.OwnerAddress, packedHash)
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

func packed(amount *big.Int, receiveAddr string, neoTxHash string) ([]byte, error) {
	packedBytes := make([]byte, 0)

	amountBytes := bytes.Repeat([]byte{0}, 32)
	aBytes := amount.Bytes()
	copy(amountBytes[len(amountBytes)-len(aBytes):], aBytes)
	packedBytes = append(packedBytes, amountBytes...)

	addr := common.HexToAddress(receiveAddr)
	packedBytes = append(packedBytes, addr.Bytes()...)

	nHash := common.HexToHash(neoTxHash)
	packedBytes = append(packedBytes, nHash.Bytes()...)
	hash := sha256.Sum256(packedBytes)

	return hash[:], nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}

func toString(b string) *pb.String {
	return &pb.String{Value: b}
}

func (d *DepositAPI) EthTransactionConfirmed(ctx context.Context, h *pb.Hash) (*pb.Boolean, error) {
	d.logger.Infof("call deposit EthTransactionConfirmed: %s", h.String())
	hash := h.GetHash()
	if hash == "" {
		return nil, errors.New("invalid hash")
	}
	confirmed, err := d.eth.HasBlockConfirmed(common.HexToHash(hash), d.cfg.EthCfg.ConfirmedHeight)
	if err != nil || !confirmed {
		d.logger.Errorf("block confirmed: %s, %t", err, confirmed)
		return toBoolean(false), fmt.Errorf("block not confirmed")
	}
	amount, address, neoTx, err := d.eth.SyncMintLog(hash)
	if err != nil {
		d.logger.Errorf("mint log: %s", err)
		return toBoolean(false), err
	}
	swapInfo, err := db.GetSwapInfoByTxHash(d.store, neoTx, types.NEO)
	if err != nil {
		return toBoolean(false), err
	}
	if swapInfo.State == types.DepositDone && swapInfo.EthTxHash != "" {
		return toBoolean(true), nil
	}

	d.logger.Infof("mint log: %d, %s, %s, neo[%s] ", amount.Uint64(), address.String(), neoTx, swapInfo.NeoTxHash)

	if amount.Int64() != swapInfo.Amount || strings.ToLower(address.String()) != strings.ToLower(swapInfo.EthUserAddr) {
		d.logger.Errorf("swap info not match: %s, amount %d, address %s", swapInfo.String(), amount.Int64(), address.String())
		return toBoolean(false), err
	}

	if swapInfo.State == types.DepositDone {
		if swapInfo.EthTxHash == "" {
			swapInfo.EthTxHash = hash
			if err := db.UpdateSwapInfo(d.store, swapInfo); err != nil {
				return toBoolean(false), err
			}
			d.logger.Infof("set deposit swap info eth hash to %s", hash)
			return toBoolean(true), nil
		} else {
			return toBoolean(true), nil
		}
	}
	if swapInfo.State == types.DepositPending {
		if err := toConfirmDepositEthTx(common.HexToHash(hash), 0, neoTx, address.String(), amount.Int64(),
			d.eth, d.cfg.EthCfg.ConfirmedHeight, d.store, d.logger, false); err != nil {
			d.logger.Errorf("deposit :%s, %s", err, hash)
			return toBoolean(false), err
		}
		return toBoolean(true), nil
	}
	return toBoolean(false), errors.New("invalid state")
}

func (d *DepositAPI) EthTransactionSent(ctx context.Context, h *pb.EthTransactionSentRequest) (*pb.Boolean, error) {
	d.logger.Infof("call deposit EthTransactionSent: %s", h.String())
	ethHash := h.GetEthTxHash()
	neoHash := h.GetNeoTxHash()
	if ethHash == "" || neoHash == "" {
		return nil, fmt.Errorf("invalid hash, %s", h)
	}
	if _, err := db.GetSwapPendingByTxEthHash(d.store, ethHash); err != nil {
		if err := db.InsertSwapPending(d.store, &types.SwapPending{
			Typ:       types.Deposit,
			EthTxHash: ethHash,
			NeoTxHash: neoHash,
		}); err != nil {
			d.logger.Error(err)
			return toBoolean(false), err
		}
	}

	go func() {
		if err := d.eth.WaitTxVerifyAndConfirmed(common.HexToHash(ethHash), 0, d.cfg.EthCfg.ConfirmedHeight); err != nil {
			d.logger.Errorf("tx confirmed: %s", err)
			return
		}
		h := pb.Hash{
			Hash: ethHash,
		}
		if _, err := d.EthTransactionConfirmed(ctx, &h); err != nil {
			d.logger.Errorf("tx confirmed: %s", err)
			return
		}
	}()
	return toBoolean(true), nil
}

func (d *DepositAPI) correctSwapPending() error {
	vTicker := time.NewTicker(4 * time.Minute)
	for {
		select {
		case <-vTicker.C:
			infos, err := db.GetSwapPendings(d.store, 0, 0)
			if err != nil {
				d.logger.Error(err)
				continue
			}
			for _, info := range infos {
				if info.Typ == types.Deposit && time.Now().Unix()-info.LastModifyTime > 60*10 {
					swapInfo, err := db.GetSwapInfoByTxHash(d.store, info.NeoTxHash, types.NEO)
					if err == nil {
						if swapInfo.State == types.DepositDone && swapInfo.EthTxHash != "" {
							_ = db.DeleteSwapPending(d.store, info)
						} else if swapInfo.State == types.DepositPending {
							d.logger.Infof("continue deposit, eth %s, neo[%s]", info.EthTxHash, swapInfo.NeoTxHash)
							if _, err := d.EthTransactionSent(context.Background(), &pb.EthTransactionSentRequest{
								EthTxHash: info.EthTxHash,
								NeoTxHash: info.NeoTxHash,
							}); err != nil {
								d.logger.Error(err)
							}
						}
					}
				}
			}
		}
	}
}

//eth transaction must fail or cancel first
func (d *DepositAPI) Refund(ctx context.Context, h *pb.Hash) (*pb.Boolean, error) {
	if d.cfg.CanRefund > 0 {
		d.logger.Infof("call deposit refund: %s", h.String())
		hash := h.GetHash()
		if hash == "" {
			return nil, fmt.Errorf("invalid hash, %s", h)
		}
		if swapInfo, err := db.GetSwapInfoByTxHash(d.store, hash, types.NEO); err == nil {
			if swapInfo.State < types.DepositDone {
				neoTx, err := d.neo.CreateUnLockTransaction(hash, swapInfo.NeoUserAddr, swapInfo.EthUserAddr, int(swapInfo.Amount), d.cfg.NEOCfg.OwnerAddress)
				if err != nil {
					d.logger.Errorf("create neo tx: %s, neo[%s]", err, hash)
					return nil, fmt.Errorf("create tx: %s", err)
				}
				d.logger.Infof("refund neo tx created: %s. neo[%s]", neoTx, hash)
				go func() {
					if _, err := d.neo.WaitTxVerifyAndConfirmed(neoTx, d.cfg.NEOCfg.ConfirmedHeight); err != nil {
						d.logger.Error(err)
						return
					}
					if _, err := d.neo.QueryLockedInfo(hash); err != nil {
						d.logger.Error(err)
						return
					}

					swapInfo.State = types.DepositRefund
					d.logger.Infof("update state to %s", types.SwapStateToString(types.DepositRefund))
					if err := db.UpdateSwapInfo(d.store, swapInfo); err != nil {
						d.logger.Error(err)
						return
					}
					d.logger.Infof("refund successfully, %s", hash)
				}()
			}
		} else {
			d.logger.Error(err)
			return nil, fmt.Errorf("neo deposit recode not found: %s", err)
		}
	}
	return toBoolean(true), nil
}

func (d *DepositAPI) EthTransactionID(ctx context.Context, hash *pb.Hash) (*pb.Hash, error) {
	neoHash := hash.GetHash()
	if neoHash == "" {
		return nil, fmt.Errorf("invalid hash, %s", neoHash)
	}
	info, err := db.GetSwapPendingByTxNeoHash(d.store, neoHash)
	if err != nil {
		return nil, err
	}
	return &pb.Hash{
		Hash: info.EthTxHash,
	}, nil
}
