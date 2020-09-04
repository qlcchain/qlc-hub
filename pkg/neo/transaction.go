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
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/pkg/log"
	u "github.com/qlcchain/qlc-hub/pkg/util"
	"go.uber.org/zap"
)

type Transaction struct {
	url          string
	client       *client.Client
	contractLE   util.Uint160
	contractAddr string
	logger       *zap.SugaredLogger
}

func NewTransaction(url, contractAddr string) (*Transaction, error) {
	c, err := client.New(context.Background(), url, client.Options{})
	if err != nil {
		return nil, err
	}
	contract, err := util.Uint160DecodeStringLE(contractAddr)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		url:          url,
		client:       c,
		contractLE:   contract,
		contractAddr: contractAddr,
		logger:       log.NewLogger("neo/transaction"),
	}, nil
}

type TransactionParam struct {
	Params   []request.Param
	Wif      string
	Sysfee   util.Fixed8
	Netfee   util.Fixed8
	ROrigin  string
	RHash    string
	FuncName string
}

func (n *Transaction) CreateTransaction(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("NewAccountFromWIF: %s", err)
	}
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
		if err = request.AddInputsAndUnspentsToTx(tx, account.Address, core.UtilityTokenID(), gas, n.client); err != nil {
			return "", fmt.Errorf("failed to add inputs and unspents to transaction: %s ", err)
		}
	} else {
		addr, err := address.StringToUint160(account.Address)
		if err != nil {
			return "", fmt.Errorf("failed to get address: %s", err)
		}
		tx.AddVerificationHash(addr)
		if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
			tx.Attributes = append(tx.Attributes, transaction.Attribute{
				Usage: transaction.Remark,
				Data:  remark(),
			})
		}
	}

	if err = account.SignTx(tx); err != nil {
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

func (n *Transaction) CreateTransactionAppendWitness(param TransactionParam) (string, error) {
	account, err := wallet.NewAccountFromWIF(param.Wif)
	if err != nil {
		return "", fmt.Errorf("new account: %s", err)
	}
	accountUint, err := address.StringToUint160(account.Address)
	if err != nil {
		return "", fmt.Errorf("address uint160: %s", err)
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

	if err := account.SignTx(tx); err != nil {
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

	n.logger.Debug(hex.EncodeToString(tx.Bytes()))
	n.logger.Debug(u.ToIndentString(tx))

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
		return nil, err
	} else if r.State != "HALT" || len(r.Stack) == 0 {
		return nil, errors.New("invalid VM state")
	}

	return StackToSwapInfo(r.Stack)
}

type SwapInfo struct {
	Amount         int64  `json:"amount"`
	UserNeoAddress string `json:"userNeoAddress"`
	State          int    `json:"state"`
	OriginText     string `json:"originText"`
	OvertimeBlocks int64  `json:"overtimeBlocks"`
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

func (n *Transaction) TxVerifyAndConfirmed(txHash string, interval int) (bool, uint32, error) {
	var txHeight uint32
	cTicker := time.NewTicker(6 * time.Second)
	cTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-cTicker.C:
			hash, err := util.Uint256DecodeStringLE(txHash)
			if err != nil {
				return false, 0, fmt.Errorf("tx verify decode hash: %s", err)
			}
			txHeight, err = n.client.GetTransactionHeight(hash)
			if err != nil {
				n.logger.Debugf("get neo tx [%s] height err: %s", hash, err)
			} else {
				goto HeightConfirmed
			}
		case <-cTimer.C:
			return false, 0, fmt.Errorf("neo tx by hash timeout: %s", txHash)
		}
	}

HeightConfirmed:
	nTicker := time.NewTicker(6 * time.Second)
	nTimer := time.NewTimer(300 * time.Second)
	for {
		select {
		case <-nTicker.C:
			nHeight, err := n.Client().GetStateHeight()
			if err != nil {
				return false, 0, err
			} else {
				nh := nHeight.BlockHeight
				n.logger.Debugf("tx [%s] current confirmed height (%d, %d)", txHash, txHeight, nh)
				if nh-txHeight >= uint32(interval) {
					return true, txHeight, nil
				}
			}
		case <-nTimer.C:
			return false, 0, fmt.Errorf("neo tx confirmed timeout: %s", txHash)
		}
	}
}

func (n *Transaction) IsLockerTimeout(txHeight uint32, interval int64) bool {
	nHeight, err := n.Client().GetStateHeight()
	if err != nil {
		return false
	}
	nh := nHeight.BlockHeight
	return nh-txHeight > uint32(interval)
}

func (n *Transaction) ValidateAddress(addr string) error {
	return n.client.ValidateAddress(addr)
}

func (n *Transaction) LockerEventFromApplicationLog(hash string) (string, State, error) {
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
	return "", 0, fmt.Errorf("can not find lock event %s", hash)
}

func (n *Transaction) CheckTxAndRHash(txHash, rHash string, confirmedHeight int, state State) (uint32, error) {
	n.logger.Infof("waiting for neo tx %s confirmed", txHash)
	b, height, err := n.TxVerifyAndConfirmed(txHash, confirmedHeight)
	if !b || err != nil {
		return 0, fmt.Errorf("neo tx confirmed: %s, %v , %s, [%s]", err, b, txHash, rHash)
	}

	rHashEvent, stateEvent, err := n.LockerEventFromApplicationLog(txHash)
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
