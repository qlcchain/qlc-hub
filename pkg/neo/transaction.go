package neo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nspcc-dev/neo-go/pkg/core"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/grpc/proto"
	"github.com/qlcchain/qlc-hub/pkg/log"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
	u "github.com/qlcchain/qlc-hub/pkg/util"
	"github.com/qlcchain/qlc-hub/signer"
)

type Transaction struct {
	signer       *signer.SignerClient
	url          string
	client       *client.Client
	contractLE   util.Uint160
	contractAddr string
	logger       *zap.SugaredLogger
}

func NewTransaction(url, contractAddr string, signer *signer.SignerClient) (*Transaction, error) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		return nil, err
	}
	contract, err := util.Uint160DecodeStringLE(contractAddr)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		signer:       signer,
		url:          url,
		client:       c,
		contractLE:   contract,
		contractAddr: contractAddr,
		logger:       log.NewLogger("neo/transaction"),
	}, nil
}

type TransactionParam struct {
	Params        []request.Param
	SignerAddress string
	Sysfee        util.Fixed8
	Netfee        util.Fixed8
	ROrigin       string
	RHash         string
	FuncName      string
	EmitIndex     string
}

func (n *Transaction) CreateTransaction(param TransactionParam) (string, error) {
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("CreateFunctionInvocationScript: %s", err)
	}

	// use library can not add remark ,may return "failed sendning tx: Block or transaction already exists and cannot be sent repeatedly"
	//re, err := n.client.SignAndPushInvocationTx(scripts, account, param.Sysfee, param.Netfee)
	//if err != nil {
	//	return "", fmt.Errorf("SignAndPushInvocationTx: %s", err)
	//}

	tx := transaction.NewInvocationTX(scripts, param.Sysfee)
	gas := param.Sysfee + param.Netfee

	if gas > 0 {
		if err = request.AddInputsAndUnspentsToTx(tx, param.SignerAddress, core.UtilityTokenID(), gas, n.client); err != nil {
			return "", fmt.Errorf("failed to add inputs and unspents to transaction: %s ", err)
		}
	} else {
		addr, err := address.StringToUint160(param.SignerAddress)
		if err != nil {
			return "", fmt.Errorf("address to unint160：%s, %s", err, param.SignerAddress)
		}
		tx.AddVerificationHash(addr)
		if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
			tx.Attributes = append(tx.Attributes, transaction.Attribute{
				Usage: transaction.Remark,
				Data:  remark(),
			})
		}
	}

	if err = n.SignTx(tx, param.SignerAddress); err != nil {
		return "", fmt.Errorf("failed to sign tx, %s", err)
	}
	txHash := tx.Hash()
	err = n.client.SendRawTransaction(tx)

	if err != nil {
		return "", fmt.Errorf("failed sendning tx: %s", err)
	}

	//n.logger.Debugf("transaction successfully: %s", re.StringLE())
	return txHash.StringLE(), nil
}

func (n *Transaction) SignTx(tx *transaction.Transaction, address string) error {
	if n.signer == nil {
		return errors.New("invalid signer")
	}
	data := tx.GetSignedPart()
	if data == nil {
		return errors.New("failed to get transaction's signed part")
	}

	if sign, err := n.signer.Sign(proto.SignType_NEO, address, data); err == nil {
		tx.Scripts = append(tx.Scripts, transaction.Witness{
			InvocationScript:   append([]byte{byte(opcode.PUSHBYTES64)}, sign.Sign...),
			VerificationScript: sign.VerifyData,
		})
		return nil
	} else {
		return err
	}
}

type witnessWrapper struct {
	transaction.Witness
	ScriptHash *util.Uint160
}

func (w witnessWrapper) GetScriptHash() *util.Uint160 {
	if w.ScriptHash == nil {
		hash := w.Witness.ScriptHash()
		w.ScriptHash = &hash
	}
	return w.ScriptHash
}

func (n *Transaction) CreateTransactionAppendWitness2(param TransactionParam) (string, error) {
	accountUint, err := address.StringToUint160(param.SignerAddress)
	if err != nil {
		return "", fmt.Errorf("addr %s to uint160: %s ", param.SignerAddress, err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("create script: %s", err)
	}
	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(accountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  n.contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  accountUint.BytesBE(),
		})
		r := remark()
		//r, _ := hex.DecodeString("00000174483c1ff2d76670ab")
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	if err := n.SignTx(tx, param.SignerAddress); err != nil {
		return "", fmt.Errorf("signTx: %s", err)
	}

	var tmp []*witnessWrapper
	// sign
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   tx.Scripts[0].InvocationScript,
			VerificationScript: tx.Scripts[0].VerificationScript,
		},
		ScriptHash: nil,
	})
	// add witness
	script := io.NewBufBinWriter()
	emit.String(script.BinWriter, "1")
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, param.EmitIndex)
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   script.Bytes(),
			VerificationScript: []byte{},
		},
		ScriptHash: &n.contractLE,
	})

	sort.Slice(tmp, func(i, j int) bool {
		h1 := tmp[i].GetScriptHash()
		h2 := tmp[j].GetScriptHash()

		return big.NewInt(0).SetBytes(h1.BytesLE()).Cmp(big.NewInt(0).SetBytes(h2.BytesLE())) < 0
	})

	size := len(tmp)
	witness := make([]transaction.Witness, size)
	for i := 0; i < size; i++ {
		witness[i] = transaction.Witness{
			InvocationScript:   tmp[i].InvocationScript,
			VerificationScript: tmp[i].VerificationScript,
		}
	}

	tx.Scripts = witness

	n.logger.Debug("tx hex: ", hex.EncodeToString(tx.Bytes()))
	n.logger.Debug("tx detail: ", u.ToIndentString(tx))

	if err := n.client.SendRawTransaction(tx); err != nil {
		return "", fmt.Errorf("sendRawTransaction: %s", err)
	}
	//n.logger.Debugf("transaction successfully: %s", tx.Hash().StringLE())
	return tx.Hash().StringLE(), nil
}

func (n *Transaction) CreateTransactionAppendWitness(param TransactionParam) (string, error) {
	accountUint, err := address.StringToUint160(param.SignerAddress)
	if err != nil {
		return "", err
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("create script: %s", err)
	}
	tx := transaction.NewInvocationTX(scripts, 0)

	// add attributes
	tx.AddVerificationHash(accountUint)
	tx.Attributes = append(tx.Attributes, transaction.Attribute{
		Usage: transaction.Script,
		Data:  n.contractLE.BytesBE(),
	})

	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Script,
			Data:  accountUint.BytesBE(),
		})
		r := remark()
		//r, _ := hex.DecodeString("00000174483c1ff2d76670ab")
		tx.Attributes = append(tx.Attributes, transaction.Attribute{
			Usage: transaction.Remark,
			Data:  r,
		})
	}

	if err := n.SignTx(tx, param.SignerAddress); err != nil {
		return "", fmt.Errorf("signTx: %s", err)
	}

	var tmp []*witnessWrapper
	// sign
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   tx.Scripts[0].InvocationScript,
			VerificationScript: tx.Scripts[0].VerificationScript,
		},
		ScriptHash: nil,
	})
	// add witness
	script := io.NewBufBinWriter()
	if param.ROrigin != "" && param.RHash == "" {
		emit.String(script.BinWriter, param.ROrigin)
	} else if param.ROrigin == "" && param.RHash != "" {
		rHex, err := hex.DecodeString(param.RHash)
		if err != nil {
			return "", fmt.Errorf("decode error: %s", err)
		}
		emit.Bytes(script.BinWriter, rHex)
	} else {
		return "", errors.New("invalid r text")
	}
	emit.Int(script.BinWriter, 1)
	emit.Opcode(script.BinWriter, opcode.PACK)
	emit.String(script.BinWriter, param.FuncName)
	tmp = append(tmp, &witnessWrapper{
		Witness: transaction.Witness{
			InvocationScript:   script.Bytes(),
			VerificationScript: []byte{},
		},
		ScriptHash: &n.contractLE,
	})

	sort.Slice(tmp, func(i, j int) bool {
		h1 := tmp[i].GetScriptHash()
		h2 := tmp[j].GetScriptHash()

		return big.NewInt(0).SetBytes(h1.BytesLE()).Cmp(big.NewInt(0).SetBytes(h2.BytesLE())) < 0
	})

	size := len(tmp)
	witness := make([]transaction.Witness, size)
	for i := 0; i < size; i++ {
		witness[i] = transaction.Witness{
			InvocationScript:   tmp[i].InvocationScript,
			VerificationScript: tmp[i].VerificationScript,
		}
	}

	tx.Scripts = witness

	n.logger.Debug("tx hex: ", hex.EncodeToString(tx.Bytes()))
	n.logger.Debug("tx detail: ", u.ToIndentString(tx))

	if err := n.client.SendRawTransaction(tx); err != nil {
		return "", fmt.Errorf("sendRawTransaction: %s", err)
	}
	//n.logger.Debugf("transaction successfully: %s", tx.Hash().StringLE())
	return tx.Hash().StringLE(), nil
}

func remark() []byte {
	remark := make([]byte, 12)
	rand.Read(remark)
	return remark
}

func (n *Transaction) Client() *client.Client {
	return n.client
}

func (n *Transaction) QuerySwapData(rHash string) (map[string]interface{}, error) {
	hash, err := hex.DecodeString(rHash)
	if err != nil {
		return nil, err
	}
	params := []smartcontract.Parameter{
		{
			Type:  smartcontract.ByteArrayType,
			Value: hash,
		},
	}
	r, err := n.client.InvokeFunction(n.contractAddr, "querySwapInfo", params, nil)
	if err != nil {
		return nil, fmt.Errorf("invoke function: %s", err)
	} else if r.State != "HALT" || len(r.Stack) == 0 {
		n.logger.Debug(hubUtil.ToString(r.Stack))
		return nil, errors.New("invalid VM state")
	}
	n.logger.Debug(hubUtil.ToString(r.Stack))
	return StackToSwapInfo(r.Stack)
}

type SwapInfo struct {
	Amount         int64  `json:"amount"`
	UserNeoAddress string `json:"userNeoAddress"`
	State          int    `json:"state"`
	OriginText     string `json:"originText"`
	OvertimeBlocks int64  `json:"overtimeBlocks"`
	TxIdIn         string `json:"txIdIn"`
	TxIdOut        string `json:"txIdOut"`
	LockedHeight   uint32 `json:"blockHeight"`
	TxIdRefund     string `json:"txIdRefund"`
	UnlockedHeight uint32 `json:"unlockedHeight"`
}

func (n *Transaction) QuerySwapInfo(rHash string) (*SwapInfo, error) {
	data, err := n.QuerySwapData(rHash)
	if err != nil {
		return nil, err
	}
	info := new(SwapInfo)
	tt, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tt, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (n *Transaction) QuerySwapInfoAndConfirmedTx(rHash string, state State, interval int) (*SwapInfo, error) {
	n.logger.Infof("querying swapInfo and confirmedTx by state %d [%s]", state, rHash)
	var sInfo *SwapInfo
	var err error
	sInfo, err = n.QuerySwapInfo(rHash)

	found := false
	if sInfo != nil {
		if sInfo.State-1 > int(state) { //swap info state from chain is begin from 1
			return nil, fmt.Errorf("get swap info is %d, not %d [%s]", sInfo.State-1, state, rHash)
		} else if sInfo.State-1 == int(state) {
			found = true
		}
	}

	if !found {
		cTicker := time.NewTicker(6 * time.Second)
		cTimer := time.NewTimer(300 * time.Second) //tx on chain may need time
		for {
			select {
			case <-cTicker.C:
				sInfo, err = n.QuerySwapInfo(rHash)
				if err == nil {
					if sInfo.State-1 > int(state) {
						return nil, fmt.Errorf("get swap info is %d, not %d [%s]", sInfo.State-1, state, rHash)
					}
					if sInfo.State-1 == int(state) {
						goto SwapFound
					}
				}
			case <-cTimer.C:
				return nil, fmt.Errorf("neo tx by hash timeout: %s, %d", rHash, state)
			}
		}
	}

SwapFound:

	n.logger.Debugf("swap info: %s", hubUtil.ToString(sInfo))
	if state == UserLock || state == WrapperLock {
		if _, err := n.TxVerifyAndConfirmed(sInfo.TxIdIn, interval); err != nil {
			return nil, err
		} else {
			return sInfo, nil
		}
	} else if state == WrapperUnlock || state == UserUnlock {
		if h, err := n.TxVerifyAndConfirmed(sInfo.TxIdOut, interval); err != nil {
			return nil, err
		} else {
			sInfo.UnlockedHeight = h
			return sInfo, nil
		}
	} else if state == RefundUser || state == RefundWrapper {
		if h, err := n.TxVerifyAndConfirmed(sInfo.TxIdRefund, interval); err != nil {
			return nil, err
		} else {
			sInfo.UnlockedHeight = h
			return sInfo, nil
		}
	} else {
		return nil, fmt.Errorf("invalid state %d, %s", state, rHash)
	}
}

func (n *Transaction) TxVerifyAndConfirmed(txHash string, interval int) (uint32, error) {
	hash, err := util.Uint256DecodeStringLE(txHash)
	if err != nil {
		return 0, fmt.Errorf("tx verify decode hash: %s, %s", err, txHash)
	}

	var txHeight uint32
	if txHeight, err = n.client.GetTransactionHeight(hash); err != nil {
		cTicker := time.NewTicker(6 * time.Second)
		cTimer := time.NewTimer(300 * time.Second) //tx on chain may need time
		for {
			select {
			case <-cTicker.C:
				txHeight, err = n.client.GetTransactionHeight(hash)
				if err != nil {
					n.logger.Debugf("get neo tx [%s] height err: %s", txHash, err)
				} else {
					goto HeightConfirmed
				}
			case <-cTimer.C:
				return 0, fmt.Errorf("tx(neo) by hash timeout: %s", txHash)
			}
		}
	}

HeightConfirmed:

	if _, _, err := n.lockerEventFromApplicationLog(txHash); err != nil { // check if failed
		return 0, err
	}

	if b, _ := n.HasConfirmedBlocksHeight(txHeight, int64(interval)); b {
		return txHeight, nil
	}
	nTicker := time.NewTicker(6 * time.Second)
	nTimer := time.NewTimer(time.Duration((interval+1)*61) * time.Second)
	for {
		select {
		case <-nTicker.C:
			if b, _ := n.HasConfirmedBlocksHeight(txHeight, int64(interval)); b {
				return txHeight, nil
			}
		case <-nTimer.C:
			return 0, fmt.Errorf("neo tx confirmed timeout: %s", txHash)
		}
	}
}

func (n *Transaction) HasConfirmedBlocksHeight(startHeight uint32, interval int64) (bool, uint32) {
	if interval < 0 {
		interval = 0
	}
	nHeight, err := n.client.GetStateHeight()
	if err != nil {
		return false, 0
	}
	nh := nHeight.BlockHeight
	n.logger.Debugf("current confirmed height (%d -> %d)", startHeight, nh)
	return nh-startHeight > uint32(interval), nh
}

func (n *Transaction) ValidateAddress(addr string) error {
	return n.client.ValidateAddress(addr)
}

func (n *Transaction) lockerEventFromApplicationLog(hash string) (string, State, error) {
	if h, err := util.Uint256DecodeStringLE(hash); err == nil {
		if l, err := n.client.GetApplicationLog(h); err == nil {
			//data, _ := json.MarshalIndent(l, "", "\t")
			//fmt.Println(string(data))
			for _, executions := range l.Executions {
				for _, events := range executions.Events {
					if events.Item.Type == smartcontract.ArrayType {
						values := events.Item.Value.([]smartcontract.Parameter)
						if len(values) == 3 {
							rHash, ok := values[1].Value.([]byte)
							if !ok {
								return "", 0, errors.New("invalid rHash value from log")
							}
							rHashStr := common.BytesToHash(rHash).String()
							if state, ok := values[2].Value.(int64); ok {
								return u.RemoveHexPrefix(rHashStr), State(state), nil
							} else {
								return u.RemoveHexPrefix(rHashStr), State(0), nil
							}
						}
					}
				}
			}
		} else {
			return "", 0, fmt.Errorf("get applicationLog: %s, %s", err, hash)
		}
	}
	return "", 0, fmt.Errorf("can not find lock event, txHash: %s", hash)
}

func (n *Transaction) CheckTxAndRHash(txHash, rHash string, confirmedHeight int, state State) (uint32, error) {
	n.logger.Infof("waiting for neo tx %s confirmed", txHash)
	height, err := n.TxVerifyAndConfirmed(txHash, confirmedHeight)
	if err != nil {
		return 0, fmt.Errorf("neo tx confirmed: %s, %s, [%s]", err, txHash, rHash)
	}

	rHashEvent, stateEvent, err := n.lockerEventFromApplicationLog(txHash)
	if err != nil {
		return 0, fmt.Errorf("neo event: %s, %s, [%s]", err, txHash, rHash)
	}
	if rHashEvent != rHash {
		return 0, fmt.Errorf("invalid rHash: %s, %s", rHashEvent, rHash)
	}
	if stateEvent != state {
		return 0, fmt.Errorf("invalid state:  %d, %d", stateEvent, state)
	}
	return height, nil
}

type State int

const (
	UserLock State = iota
	WrapperUnlock
	RefundUser
	WrapperLock
	UserUnlock
	RefundWrapper
)

func (n *Transaction) GetTransactionHeight(txHash string) (uint32, error) {
	hash, err := util.Uint256DecodeStringLE(txHash)
	if err != nil {
		return 0, fmt.Errorf("tx verify decode hash: %s", err)
	}
	return n.client.GetTransactionHeight(hash)
}

func (n *Transaction) Balance(addr string, asset string) (int64, error) {
	address, err := address.StringToUint160(addr)
	if err != nil {
		return 0, err
	}
	r, err := n.client.GetNEP5Balances(address)
	if err != nil {
		return 0, nil
	}

	for _, assetBalance := range r.Balances {
		if assetBalance.Asset.StringLE() == asset {
			amount, err := strconv.ParseInt(assetBalance.Amount, 10, 64)
			if err != nil {
				return 0, nil
			}
			return amount, nil
		}
	}
	return 0, nil
}

func (n *Transaction) SignData(address string, str string) (*proto.SignResponse, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return n.signer.Sign(proto.SignType_NEO, address, data)
}
