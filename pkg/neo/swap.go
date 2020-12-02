package neo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"

	"github.com/nspcc-dev/neo-go/pkg/core"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/qlcchain/qlc-hub/grpc/proto"
	hubUtil "github.com/qlcchain/qlc-hub/pkg/util"
)

// get unsigned lock tx
func (n *Transaction) UnsignedLockTransaction(nep5SenderAddr, erc20ReceiverAddr string, amount int) (string, string, error) {
	params := []request.Param{
		FunctionName("lock"),
		ArrayParams([]request.Param{
			AddressParam(nep5SenderAddr),
			IntegerTypeParam(amount),
			ArrayTypeParam(hubUtil.RemoveHexPrefix(erc20ReceiverAddr)),
		}),
	}
	return n.createUnsignedTransaction(TransactionParam{
		Params:        params,
		SignerAddress: nep5SenderAddr,
	})
}

func (n *Transaction) createUnsignedTransaction(param TransactionParam) (string, string, error) {
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", "", fmt.Errorf("CreateFunctionInvocationScript: %s", err)
	}

	tx := transaction.NewInvocationTX(scripts, param.Sysfee)
	gas := param.Sysfee + param.Netfee

	if gas > 0 {
		if err = request.AddInputsAndUnspentsToTx(tx, param.SignerAddress, core.UtilityTokenID(), gas, n.client); err != nil {
			return "", "", fmt.Errorf("failed to add inputs and unspents to transaction: %s ", err)
		}
	} else {
		addr, err := address.StringToUint160(param.SignerAddress)
		if err != nil {
			return "", "", fmt.Errorf("address to unint160：%s, %s", err, param.SignerAddress)
		}
		tx.AddVerificationHash(addr)
		if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
			tx.Attributes = append(tx.Attributes, transaction.Attribute{
				Usage: transaction.Remark,
				Data:  remark(),
			})
		}
	}
	data := tx.GetSignedPart()
	if data == nil {
		return "", "", errors.New("failed to get transaction's signed part")
	}

	n.pendingTx.Store(tx.Hash().StringLE(), tx)
	return tx.Hash().StringLE(), hex.EncodeToString(data), nil
}

// send lock tx
func (n *Transaction) SendLockTransaction(txHash, signature, publicKey, address string) (string, error) {
	txObj, ok := n.pendingTx.Load(txHash)
	if !ok {
		return "", fmt.Errorf("tx not found: %s", txHash)
	}
	tx, ok := txObj.(*transaction.Transaction)
	if !ok {
		return "", fmt.Errorf("invalid tx : %s", txHash)
	}
	verificationScript, signatureBytes, err := signatureVerify(tx.GetSignedPart(), signature, publicKey, address)
	if err != nil {
		return "", err
	}
	tx.Scripts = append(tx.Scripts, transaction.Witness{
		InvocationScript:   append([]byte{byte(opcode.PUSHBYTES64)}, signatureBytes...),
		VerificationScript: verificationScript,
	})
	c := n.Client()
	if c == nil {
		return "", errors.New("invalid neo endpoints")
	}
	err = c.SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("failed sendning tx: %s", err)
	}
	return tx.Hash().StringLE(), nil
}

func signatureVerify(unsignedData []byte, signature, publicKey, address string) ([]byte, []byte, error) {
	pk, err := keys.NewPublicKeyFromString(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid publickey: %s", publicKey)
	}
	signBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid signature: %s", signature)
	}
	if pk.Address() != address {
		return nil, nil, fmt.Errorf("invaild publickey, publickey:%s, address:%s", publicKey, address)
	}
	hash := sha256.Sum256(unsignedData)
	if !pk.Verify(signBytes, hash[:]) {
		return nil, nil, fmt.Errorf("invaild signature, unsignedData:%s, signature:%s", hex.EncodeToString(unsignedData), signature)
	}
	return pk.GetVerificationScript(), signBytes, nil
}

// create lock tx
func (n *Transaction) CreateLockTransaction(nep5SenderAddr, erc20ReceiverAddr, wif string, amount int) (string, error) {
	params := []request.Param{
		FunctionName("lock"),
		ArrayParams([]request.Param{
			AddressParam(nep5SenderAddr),
			IntegerTypeParam(amount),
			ArrayTypeParam(hubUtil.RemoveHexPrefix(erc20ReceiverAddr)),
		}),
	}
	r, err := n.createTransaction(TransactionParam{
		Params: params,
	}, wif)
	if err != nil {
		return "", fmt.Errorf("lock/CreateTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) createTransaction(param TransactionParam, wif string) (string, error) {
	account, err := wallet.NewAccountFromWIF(wif)
	if err != nil {
		return "", fmt.Errorf("NewAccountFromWIF: %s", err)
	}
	scripts, err := request.CreateFunctionInvocationScript(n.contractLE, param.Params)
	if err != nil {
		return "", fmt.Errorf("CreateFunctionInvocationScript: %s", err)
	}

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

// create unlock tx
func (n *Transaction) CreateUnLockTransaction(ethTxId, nep5ReceiverAddr, erc20SenderAddr string, amount int, signerAddress string) (string, error) {
	params := []request.Param{
		FunctionName("unlock"),
		ArrayParams([]request.Param{
			AddressParam(nep5ReceiverAddr),
			IntegerTypeParam(amount),
			ArrayTypeParam(hubUtil.RemoveHexPrefix(erc20SenderAddr)),
			ArrayTypeParam(hubUtil.RemoveHexPrefix(ethTxId)),
		}),
	}
	r, err := n.createTransactionAppendWitness(TransactionParam{
		Params:        params,
		SignerAddress: signerAddress,
		FuncName:      "unlockV2",
		EmitIndex:     "1",
	})
	if err != nil {
		return "", fmt.Errorf("unlock/createTransaction: %s", err)
	}
	return r, nil
}

func (n *Transaction) createTransactionAppendWitness(param TransactionParam) (string, error) {
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
		h1 := tmp[i].getScriptHash()
		h2 := tmp[j].getScriptHash()

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
	n.logger.Debug("tx detail: ", hubUtil.ToIndentString(tx))

	c := n.Client()
	if c == nil {
		return "", errors.New("invalid neo endpoints")
	}

	if err := c.SendRawTransaction(tx); err != nil {
		return "", fmt.Errorf("sendRawTransaction: %s", err)
	}
	//n.logger.Debugf("transaction successfully: %s", tx.Hash().StringLE())
	return tx.Hash().StringLE(), nil
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

func remark() []byte {
	remark := make([]byte, 12)
	rand.Read(remark)
	return remark
}

type witnessWrapper struct {
	transaction.Witness
	ScriptHash *util.Uint160
}

func (w witnessWrapper) getScriptHash() *util.Uint160 {
	if w.ScriptHash == nil {
		hash := w.Witness.ScriptHash()
		w.ScriptHash = &hash
	}
	return w.ScriptHash
}

type LockedInfo struct {
	FromAddress    string `json:"fromAddress"`
	ToAddress      string `json:"toAddress"`
	Txid           string `json:"txid"`
	UserEthAddress string `json:"userEthAddress"`
	Amount         int64  `json:"amount"`
	Timestamp      int64  `json:"timestamp"`
	BlockHeight    int64  `json:"blockHeight"`
	Typ            int    `json:"typ"`
}

func (l *LockedInfo) String() string {
	bs, _ := json.Marshal(l)
	return string(bs)
}

func (n *Transaction) QueryLockedInfo(hash string) (*LockedInfo, error) {
	data, err := n.QuerySwapData(hubUtil.RemoveHexPrefix(hash))
	if err != nil {
		return nil, fmt.Errorf("%s, %s", err, hash)
	}
	info := new(LockedInfo)
	tt, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tt, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (n *Transaction) QuerySwapData(h string) (map[string]interface{}, error) {
	hash, err := hex.DecodeString(h)
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

func (n *Transaction) SwapEnd(hash string) {
	n.pendingTx.Delete(hash)
}
