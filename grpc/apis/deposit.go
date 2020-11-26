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
	return &DepositAPI{
		cfg:    cfg,
		neo:    neo,
		eth:    e,
		ctx:    ctx,
		store:  s,
		signer: signer,
		logger: log.NewLogger("api/deposit"),
	}
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
	address := request.GetAddress()
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
	neoInfo, err := d.neo.QuerySwapInfo(hash.StringBE())
	if err != nil {
		return err
	}
	d.logger.Infof("neo transaction verify successfully. neo[%s]", neoTxHash)

	swapInfo := &types.SwapInfo{
		State:       types.DepositPending,
		Amount:      neoInfo.Amount,
		EthTxHash:   "",
		NeoTxHash:   neoTxHash,
		EthUserAddr: neoInfo.UserEthAddress,
		NeoUserAddr: neoInfo.FromAddress,
		StartTime:   time.Now().Unix(),
	}
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

func (d *DepositAPI) GetEthOwnerSign(ctx context.Context, request *proto.EthOwnerSignRequest) (*proto.String, error) {
	d.logger.Infof("call deposit GetEthOwnerSign: %s", request.String())
	neoTxHash := request.GetNeoTxHash()

	swapInfo, err := db.GetSwapInfoByTxHash(d.store, neoTxHash, types.NEO)
	if err != nil {
		d.logger.Errorf("neo not locked, neo tx[%s]", neoTxHash)
		return nil, fmt.Errorf("neo not locked")
	}
	if swapInfo.State == types.DepositDone {
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
