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
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	pb "github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"github.com/qlcchain/qlc-hub/pkg/neo"
	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
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
	d.logger.Infof("deposit GetNeoUnsignedData: %s", request.String())
	receiveAddr := request.GetErc20ReceiveAddr()
	amount := request.GetAmount()
	if receiveAddr == "" || amount <= 0 {
		d.logger.Error("unsigned invalid params")
		return nil, errors.New("invalid params")
	}

	//todo debug
	userAddress := "ARmZ7hzU1SapXr5p75MC8Hh9xSMRStM4JK"
	assetsAddr := "Ac2EMY7wCV9Hn9LR1wMWbjgGCqtVofmd6W"
	_, rHash := util.Sha256Hash()
	txHash, data, err := d.neo.UnsignedLockTransaction(userAddress, assetsAddr, rHash, int(amount))
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}
	d.logger.Infof("unsigned tx, hash:%s, data:%s", txHash, data)
	return &pb.PackNeoTxResponse{
		TxHash:       txHash,
		UnsignedData: data,
	}, nil
}

func (d *DepositAPI) SendNeoTransaction(ctx context.Context, request *pb.NeoTransactionRequest) (*pb.String, error) {
	d.logger.Infof("deposit SendNeoTransaction: %s", request.String())

	neoTxHash := request.GetTxHash()
	signature := request.GetSignature()
	publicKey := request.GetPublicKey()
	address := request.GetAddress()
	if neoTxHash == "" || signature == "" || publicKey == "" || address == "" {
		d.logger.Error("transaction invalid params")
		return nil, errors.New("invalid params")
	}

	tx, err := d.neo.SendLockTransaction(neoTxHash, signature, publicKey, address)
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}

	go func() {
		_, err := d.neo.WaitTxVerifyAndConfirmed(neoTxHash, d.cfg.NEOCfg.ConfirmedHeight)
		if err != nil {
			d.logger.Error(err)
			return
		}
		//todo check neo swap info
		swapInfo := &types.SwapInfo{
			State:       types.DepositPending,
			Amount:      0,
			EthTxHash:   "",
			NeoTxHash:   tx,
			EthUserAddr: "",
			NeoUserAddr: address,
			StartTime:   time.Now().Unix(),
		}
		if err := db.InsertSwapInfo(d.store, swapInfo); err != nil {
			d.logger.Error(err)
		}
	}()
	return toString(tx), nil
}

func (d *DepositAPI) GetEthOwnerSign(ctx context.Context, request *proto.EthOwnerSignRequest) (*proto.String, error) {
	d.logger.Infof("deposit GetEthOwnerSign: %s", request.String())
	amount := request.GetAmount()
	receiveAddr := request.GetReceiveAddr()
	neoTxHash := request.GetNeoTxHash()

	if neoTxHash == "" || receiveAddr == "" || amount == 0 {
		d.logger.Error("sign invalid params")
		return nil, errors.New("invalid params")
	}

	if err := d.verifyNeoTx(neoTxHash, amount, receiveAddr); err != nil {
		d.logger.Error(err)
		return nil, err
	}

	sign, err := d.signData(big.NewInt(amount), receiveAddr, util.RemoveHexPrefix(neoTxHash))
	if err != nil {
		d.logger.Error(err)
		return nil, err
	}

	return toString(sign), nil
}

func (d *DepositAPI) verifyNeoTx(neoTxHash string, amount int64, receiverAddr string) error {
	if err := d.neo.HasTransactionConfirmed(neoTxHash, d.cfg.NEOCfg.ConfirmedHeight); err != nil {
		return err
	}
	//todo Query neo swap info tx
	return nil
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

	nBytes, err := hex.DecodeString(neoTxHash)
	if err != nil {
		return nil, err
	}
	packedBytes = append(packedBytes, nBytes...)
	hash := sha256.Sum256(packedBytes)

	return hash[:], nil
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}

func toString(b string) *pb.String {
	return &pb.String{Value: b}
}
