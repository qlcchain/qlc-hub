package wrapper

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	qlctypes "github.com/qlcchain/qlc-go-sdk/pkg/types"
	"github.com/qlcchain/qlc-hub/common/types"
	"github.com/qlcchain/qlc-hub/log"
	"github.com/qlcchain/qlc-hub/util/request"
	"github.com/x-contract/neo-go-sdk/neocliapi"
	"github.com/x-contract/neo-go-sdk/neotransaction"
	"github.com/x-contract/neo-go-sdk/neoutils"
	"go.uber.org/zap"
)

var (
	neoEndPoint = "http://seed1.ngd.network:20332"
)

var (
	OpLock              = "lock"
	OpUnLock            = "unlock"
	OpWrapperLock       = "wrapperLock"
	OpWrapperUnlock     = "wrapperUnlock"
	OpRefundWrapper     = "refundWrapper"
	OpGetLockInfo       = "getLockInfo"
	OpGetApplicationLog = "getapplicationlog"
	OpQuerySwapInfo     = "querySwapInfo"
)

const (
	Nep5ActionUserLock      int64 = 0
	Nep5ActionWrapperUnlock       = 1
	Nep5ActionRefundUser          = 2
	Nep5ActionWrapperLock         = 3
	Nep5ActionUserUnlock          = 4
	Nep5ActionRefundWrapper       = 5
)

type Transaction struct {
	neoNode      string
	contract     string
	officialNode []string
	logger       *zap.SugaredLogger
}

type NeoTxNotification struct {
	Amount   int64
	Value    int64
	Action   string
	Fromaddr string
	Toaddr   string
	HashKey  string
}

func NewTransaction(neoNode string, contract string, officialNode []string) *Transaction {
	var ctstr string
	if contract[0] == '0' && (contract[1] == 'x' || contract[1] == 'X') {
		copy([]byte(ctstr), string(contract[2:]))
	} else {
		ctstr = contract
	}
	//fmt.Println("NewTransaction contract:",ctstr)
	return &Transaction{
		neoNode:      neoNode,
		contract:     ctstr,
		officialNode: officialNode,
		logger:       log.NewLogger("transaction")}
}

func (t *Transaction) Invoke(operation string, args []interface{}, wif string) (string, error) {
	script, err := t.CreateScript(operation, args, true)
	if err != nil {
		t.logger.Errorf("create script error: %s", err)
		return "", err
	}
	id, err := t.CreateTransaction(script, wif)
	if err != nil {
		t.logger.Errorf("create transaction error: %s", err)
		return "", err
	}
	return id, nil
}

func (t *Transaction) CreateScript(operation string, args []interface{}, withNonce bool) ([]byte, error) {
	contractHash, err := hex.DecodeString(t.contract)
	if err != nil {
		t.logger.Errorf("CreateScript bad contract：%", t.contract)
		return nil, err
	}
	return neotransaction.BuildCallMethodScript(contractHash, operation, args, withNonce)
}

func (t *Transaction) CreateTransaction(script []byte, wif string) (string, error) {
	tx := neotransaction.CreateInvocationTransaction()
	extra := tx.ExtraData.(*neotransaction.InvocationExtraData)
	extra.Script = script
	// Perhaps the transaction need Witness
	if wif != "" {
		key, err := neotransaction.DecodeFromWif(wif)
		if err != nil {
			t.logger.Error(err)
			return "", err
		}
		tx.AppendAttribute(neotransaction.UsageScript, key.CreateBasicAddress().ScripHash)
		tx.AppendBasicSignWitness(key)
	}

	//t.logger.Debugf("CreateTransaction txID: ", tx.TXID())
	rawtx := tx.RawTransactionString()
	b := neocliapi.SendRawTransaction(t.neoNode, rawtx)
	if !b {
		return "", errors.New("sendRawTransaction fail")
	}
	return tx.TXID(), nil
}

func (t *Transaction) CreateTransactionWithAttr(script []byte, wif string, operation string, args []interface{}) (string, error) {
	tx := neotransaction.CreateInvocationTransaction()
	extra := tx.ExtraData.(*neotransaction.InvocationExtraData)
	extra.Script = script
	// Perhaps the transaction need Witness
	if wif != "" {
		key, err := neotransaction.DecodeFromWif(wif)
		if err != nil {
			t.logger.Error(err)
			return "", err
		}

		tx.AppendAttribute(neotransaction.UsageScript, key.CreateBasicAddress().ScripHash)
		tx.AppendBasicSignWitness(key)
	}
	contractHash, err := hex.DecodeString(t.contract)
	if err != nil {
		t.logger.Errorf("CreateScript bad contract：%", t.contract)
		return "", err
	}
	tx.AppendAttribute(neotransaction.UsageScript, contractHash)
	contractscript, err := t.BuildContractVerifyScript(operation, args)
	if err != nil {
		t.logger.Errorf("BuildContractVerifyScript faield")
		return "", err
	}
	tx.AppendWitness(contractscript)
	rawtx := tx.RawTransactionString()
	t.logger.Debugf("get rawtx(%s)", rawtx)
	b := neocliapi.SendRawTransaction(t.neoNode, rawtx)
	if !b {
		t.logger.Errorf("SendRawTransaction faield")
		return "", errors.New("sendRawTransaction fail")
	}
	return tx.TXID(), nil
}

func (t *Transaction) GetApplicationLog(txID string) (interface{}, error) {
	para := []interface{}{txID}
	result, err := request.HttpRequest(OpGetApplicationLog, para, t.neoNode)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Transaction) WaitApplicationLog(txID string) (interface{}, error) {
	time.Sleep(15 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 30; i++ {
		<-ticker.C
		r, err := t.GetApplicationLog(txID)
		if err == nil {
			return r, nil
		}
	}
	return nil, errors.New("get application log time out")
}

func (t *Transaction) GetLockInfoByTransaction(txid string) (*types.LockInfo, error) {
	b, err := hex.DecodeString(txid)
	if err != nil {
		return nil, err
	}
	args := []interface{}{neoutils.Reverse(b)}
	fmt.Println(hex.EncodeToString(neoutils.Reverse(b)))
	id, err := t.Invoke(OpGetLockInfo, args, "")
	if err != nil {
		return nil, err
	}
	i, err := t.WaitApplicationLog(id)
	if err != nil {
		return nil, err
	}
	return t.ParseTransactionResult(i)
}

func (t *Transaction) GetLockInfoByTxid(txid string) (*types.LockInfo, error) {
	i, err := t.GetApplicationLog(txid)
	if err != nil {
		return nil, err
	}
	return t.ParseTransactionResult(i)
}

func (t *Transaction) GetLockInfoByInvokeFunctionWithOfficialNode(txid string) (*types.LockInfo, error) {
	info, err := t.GetLockInfoByInvokeFunction(txid, t.neoNode)
	if err == nil {
		t.logger.Infof("get lock info from %s: %s %s", t.neoNode, txid, info.String())
		return info, nil
	}
	t.logger.Infof("can not get lock info by node: %s, %s", txid, t.neoNode)
	if t.officialNode == nil || len(t.officialNode) == 0 {
		return nil, errors.New("official node is null")
	}
	for _, node := range t.officialNode {
		info, err = t.GetLockInfoByInvokeFunction(txid, node)
		if err == nil {
			t.logger.Infof("get lock info from %s: %s %s", node, txid, info.String())
			return info, nil
		}
		t.logger.Infof("can not get lock info by node from officialNode: %s, %s", txid, node)
	}
	return nil, fmt.Errorf("get lockinfo error: %s", err)
}

func (t *Transaction) GetLockInfoByInvokeFunction(txid string, neoNode string) (*types.LockInfo, error) {
	b, err := hex.DecodeString(txid)
	if err != nil {
		return nil, err
	}
	p := make([]map[string]string, 0)
	m := make(map[string]string)
	m["type"] = "ByteArray"
	m["value"] = hex.EncodeToString(neoutils.Reverse(b))
	p = append(p, m)
	param := []interface{}{
		t.contract,
		"getLockInfo",
		p,
	}

	i, err := request.HttpRequest("invokefunction", param, neoNode)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}
	return t.ParseTransactionResult(i)
}

func (t *Transaction) UnLock(txid string, wif string) (string, error) {
	b, err := hex.DecodeString(txid)
	if err != nil {
		return "", err
	}
	args := []interface{}{neoutils.Reverse(b)}
	script, err := t.CreateScript(OpUnLock, args, true)
	if err != nil {
		return "", err
	}
	id, err := t.CreateTransaction(script, wif)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (t *Transaction) ParseTransactionResult(i interface{}) (*types.LockInfo, error) {
	m, err := parseTransactionStack(i)
	if err != nil {
		return nil, err
	}
	lockinfo, err := parseLockInfo(m)
	if err != nil {
		return nil, err
	}
	return lockinfo, nil
}

func parseTransactionStack(i interface{}) (map[string]interface{}, error) {
	r, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transaction data error: %s\n", i)
	}
	var stack interface{}
	executions, ok := r["executions"]
	if !ok {
		stack, ok = r["stack"]
		if !ok {
			return nil, errors.New("transaction data has no stack")
		}
	} else {
		es, ok := executions.([]interface{})
		if !ok || len(es) < 1 {
			return nil, errors.New("executions data not found")
		}
		e, ok := es[0].(map[string]interface{})
		if !ok {
			return nil, errors.New("executions data error")
		}
		stack, ok = e["stack"]
		if !ok {
			return nil, errors.New("transaction data has no stack")
		}
	}

	vs, ok := stack.([]interface{})
	if !ok || len(vs) < 1 {
		return nil, errors.New("stack data error")
	}
	v, ok := vs[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("value data error")
	}
	return v, nil
}

func parseLockInfo(i map[string]interface{}) (*types.LockInfo, error) {
	if v, ok := i["value"]; ok {
		if vs, ok := v.([]interface{}); ok && len(vs) == 7 {
			var lockInfo types.LockInfo
			// NeoAddress
			value, err := getValue(vs[0])
			if err != nil {
				return nil, err
			}
			b, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			address, err := neotransaction.ParseAddressHash(neoutils.HASH160(b))
			if err != nil {
				return nil, err
			}
			lockInfo.NeoAddress = address.GetAddrString()

			// MultiSigAddress
			value, err = getValue(vs[1])
			if err != nil {
				return nil, err
			}
			b, err = hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			address, err = neotransaction.ParseAddressHash(neoutils.HASH160(b))
			if err != nil {
				return nil, err
			}
			lockInfo.MultiSigAddress = address.GetAddrString()

			// QlcAddress
			if value, err = getValue(vs[2]); err != nil {
				return nil, err
			}
			bs, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			if lockInfo.QlcAddress, err = qlctypes.HexToAddress(string(bs)); err != nil {
				return nil, err
			}

			// LockTimestamp
			if value, err = getValue(vs[3]); err != nil {
				return nil, err
			}
			if lockInfo.LockTimestamp, err = strconv.ParseInt(value, 10, 64); err != nil {
				return nil, err
			}

			// UnLockTimestamp
			if value, err = getValue(vs[4]); err != nil {
				return nil, err
			}
			if lockInfo.UnLockTimestamp, err = strconv.ParseInt(value, 10, 64); err != nil {
				return nil, err
			}

			// Amount
			if value, err = getValue(vs[5]); err != nil {
				return nil, err
			}
			if len(value) < 16 {
				value = fmt.Sprintf("%s%s", value, strings.Repeat("0", 16-len(value)))
			}
			d, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			u := binary.LittleEndian.Uint64(d)
			lockInfo.Amount = qlctypes.Balance{Int: big.NewInt(int64(u))}

			// State
			if value, err = getValue(vs[6]); err != nil {
				return nil, err
			}
			if value == "" {
				lockInfo.State = true
			} else {
				lockInfo.State = false
			}
			return &lockInfo, nil
		}
		return nil, fmt.Errorf("value is not lockinfo struct : %s", i)
	}
	return nil, fmt.Errorf("data not found: %s ", i)
}

func getValue(i interface{}) (string, error) {
	if m, ok := i.(map[string]interface{}); ok {
		if v, ok := m["value"]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", errors.New("value not correct")
		}
		return "", errors.New("value not found")
	}
	return "", errors.New("error value pattern ")
}

func (t *Transaction) Nep5ContractWrapperLock(amount, locknum int64, uaddr, lockhash string) (string, error) {
	t.logger.Debugf("Nep5ContractWrapperLock  amount(%d) locknum(%d) uaddr(%s) lockhash(%s)", amount, locknum, uaddr, lockhash)
	lh, err := hex.DecodeString(lockhash)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperLock: lockhash decode err  %s", lockhash)
		return "", err
	}
	wrapperkey, err := neotransaction.DecodeFromWif(WrapperNeoPrikey)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperLock: WrapperNeoAccount WrapperNeoPrikey err  %s", WrapperNeoPrikey)
		return "", err
	}
	wapaccount := wrapperkey.CreateBasicAddress()
	//amountWei := amount * WrapperGasWeiNum
	amountWei := amount //这里是因为从eth监听到的amount已经是转过的了
	//param := []interface{}{neoutils.Reverse(lh),wapaccount.ScripHash,amountWei,uaccount.ScripHash,locknum}
	param := []interface{}{lh, wapaccount.ScripHash, amountWei, uaddr, locknum}
	//t.logger.Debugf("Nep5ContractWrapperLock  uaddr(%d:%v) wapAccount(%d:%v)", len(uaddr), uaddr, len(wapaccount.ScripHash), wapaccount.ScripHash)
	script, err := t.CreateScript(OpWrapperLock, param, true)
	if err != nil {
		return "", err
	}
	//t.logger.Debugf("Nep5ContractWrapperLock: script %s,prikey %s",script,WrapperNeoPrikey)
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		return "", err
	}
	return id, nil
}

//Nep5ContractQuerySwapInfo
func (t *Transaction) Nep5ContractQuerySwapInfo(lockhash string) (string, error) {
	lh, err := hex.DecodeString(lockhash)
	if err != nil {
		t.logger.Error("Nep5ContractQuerySwapInfo: lockhash decode err  %s", lockhash)
		return "", err
	}
	wrapperkey, err := neotransaction.DecodeFromWif(WrapperNeoPrikey)
	if err != nil {
		t.logger.Error("Nep5ContractQuerySwapInfo: WrapperNeoAccount WrapperNeoPrikey err  %s", WrapperNeoPrikey)
		return "", err
	}
	wapaccount := wrapperkey.CreateBasicAddress()
	param := []interface{}{lh, wapaccount.ScripHash}
	script, err := t.CreateScript(OpWrapperLock, param, true)
	if err != nil {
		return "", err
	}
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5ContractWrapperUnlock(locksource, addr string) (string, error) {
	//args := []interface{}{neoutils.Reverse([]byte(locksource)),addr}
	args := []interface{}{locksource, addr}
	script, err := t.CreateScript(OpWrapperUnlock, args, true)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperUnlock:failed", err)
		return "", err
	}
	//id, err := t.CreateTransactionWithAttr(script, WrapperNeoPrikey, OpWrapperUnlock, args)
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		t.logger.Error("CreateTransaction:failed", err)
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5ContractWrapperRefund(locksource string) (string, error) {
	wrapperkey, err := neotransaction.DecodeFromWif(WrapperNeoPrikey)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperRefund: WrapperNeoAccount WrapperNeoPrikey err  %s", WrapperNeoPrikey)
		return "", err
	}
	wapaccount := wrapperkey.CreateBasicAddress()
	//args := []interface{}{neoutils.Reverse([]byte(locksource)),wapaccount.ScripHash}
	args := []interface{}{locksource, wapaccount.ScripHash}
	script, err := t.CreateScript(OpRefundWrapper, args, true)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperRefund: CreateScripterr  %s", err)
		return "", err
	}
	//id, err := t.CreateTransactionWithAttr(script, WrapperNeoPrikey, OpRefundWrapper, args)
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		t.logger.Error("Nep5ContractWrapperRefund: CreateTransactio  %s", err)
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5TransactionVerify(txid string) (status int, err error) {
	_, err = t.WaitApplicationLog(txid)
	if err != nil {
		return CchTxVerifyStatFailed, err
	}
	return CchTxVerifyStatOk, nil
}

func (t *Transaction) parseTxNotification(i map[string]interface{}) (*NeoTxNotification, error) {
	if v, ok := i["value"]; ok {
		if vs, ok := v.([]interface{}); ok && len(vs) == 4 {
			var txInfo NeoTxNotification
			// action
			value, err := getValue(vs[0])
			if err != nil {
				return nil, err
			}
			b, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			txInfo.Action = string(b)

			// fromaddr
			value, err = getValue(vs[1])
			if err != nil {
				return nil, err
			}
			b, err = hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			faddr, err := neotransaction.ParseAddressHash(neoutils.HASH160(b))
			if err != nil {
				return nil, err
			}
			txInfo.Fromaddr = faddr.GetAddrString()

			// toaddr
			if value, err = getValue(vs[2]); err != nil {
				return nil, err
			}
			c, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			taddr, err := neotransaction.ParseAddressHash(neoutils.HASH160(c))
			if err != nil {
				return nil, err
			}
			txInfo.Toaddr = taddr.GetAddrString()

			// Amount
			if value, err = getValue(vs[3]); err != nil {
				return nil, err
			}
			if len(value) < 16 {
				value = fmt.Sprintf("%s%s", value, strings.Repeat("0", 16-len(value)))
			}
			d, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			u := binary.LittleEndian.Uint64(d)
			txInfo.Amount = int64(u)
			return &txInfo, nil
		}
		return nil, fmt.Errorf("value is not txInfo struct : %s", i)
	}
	return nil, fmt.Errorf("data not found: %s ", i)
}

func (t *Transaction) parseTransactionNotification(i interface{}) (*NeoTxNotification, error) {
	txInfo := new(NeoTxNotification)
	r, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transaction data error: %s", i)
	}
	executions, ok := r["executions"]
	if !ok {
		return nil, errors.New("transaction data has no executions")
	}
	es, ok := executions.([]interface{})
	if !ok || len(es) < 1 {
		return nil, errors.New("executions data not found")
	}
	e, ok := es[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("executions data error")
	}
	notifications, ok := e["notifications"]
	if !ok {
		return nil, errors.New("transaction data has no notifications")
	}
	na, ok := notifications.([]interface{})
	//t.logger.Debugf("get notifications %v ", na)
	if !ok || len(na) < 1 {
		return nil, errors.New("notifications data not found")
	}
	n1, ok := na[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("notifications data1 error")
	}
	state1, ok := n1["state"]
	if !ok {
		return nil, errors.New("transaction data has no state")
	}
	vs1, ok := state1.(map[string]interface{})
	if !ok || len(vs1) < 1 {
		return nil, errors.New("value data error")
	}
	//t.logger.Debugf("get v1(%v)",vs1)
	if v1, ok := vs1["value"]; ok {
		if vs1, ok := v1.([]interface{}); ok && len(vs1) == 4 {
			// action
			value, err := getValue(vs1[0])
			if err != nil {
				return nil, err
			}
			b, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			txInfo.Action = string(b)
			// fromaddr
			value, err = getValue(vs1[1])
			if err != nil {
				return nil, err
			}
			b, err = hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			faddr, err := neotransaction.ParseAddressHash(neoutils.HASH160(b))
			if err != nil {
				return nil, err
			}
			txInfo.Fromaddr = faddr.GetAddrString()

			// toaddr
			if value, err = getValue(vs1[2]); err != nil {
				return nil, err
			}
			c, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			taddr, err := neotransaction.ParseAddressHash(neoutils.HASH160(c))
			if err != nil {
				return nil, err
			}
			txInfo.Toaddr = taddr.GetAddrString()

			// Amount
			if value, err = getValue(vs1[3]); err != nil {
				return nil, err
			}
			if len(value) < 16 {
				value = fmt.Sprintf("%s%s", value, strings.Repeat("0", 16-len(value)))
			}
			d, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			u := binary.LittleEndian.Uint64(d)
			txInfo.Amount = int64(u)
		}
	}

	n2, ok := na[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("notifications data1 error")
	}
	state2, ok := n2["state"]
	if !ok {
		return nil, errors.New("state data2 has no state")
	}
	vs2, ok := state2.(map[string]interface{})
	if !ok || len(vs2) < 1 {
		return nil, errors.New("value vs2 error")
	}
	//t.logger.Debugf("get v2(%v)",vs2)
	if v2, ok := vs2["value"]; ok {
		if vs2, ok := v2.([]interface{}); ok && len(vs2) == 3 {
			// action
			value, err := getValue(vs2[0])
			if err != nil {
				return nil, err
			}
			b, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			txInfo.Action = string(b)
			// hashkey
			value, err = getValue(vs2[1])
			if err != nil {
				return nil, err
			}
			txInfo.HashKey = value

			// Value
			if value, err = getValue(vs2[2]); err != nil {
				return nil, err
			}
			txInfo.Value, _ = strconv.ParseInt(value, 10, 64)
		}
	}
	t.logger.Debugf("get txInfo(%d:%d:%s:%s:%s:%s)", txInfo.Value, txInfo.Amount, txInfo.Action, txInfo.Fromaddr, txInfo.Toaddr, txInfo.HashKey)
	return txInfo, nil
}

func (t *Transaction) Nep5GetTxInfo(txid string) (*NeoTxNotification, error) {
	param := []interface{}{txid}
	i, err := request.HttpRequest("getapplicationlog", param, neoEndPoint)
	if err != nil {
		return nil, err
	}
	return t.parseTransactionNotification(i)
}

func (t *Transaction) Nep5VerifyByTxid(event *EventInfo) (ret int, err error) {
	var txid string
	switch event.Status {
	case cchNep5MortgageStatusWaitNeoLockVerify:
		txid = event.NeoLockTxhash
	case cchNep5MortgageStatusWaitNeoUnlockVerify:
		txid = event.NeoUnlockTxhash
	case cchEthRedemptionStatusWaitNeoLockVerify:
		txid = event.NeoLockTxhash
	case cchEthRedemptionStatusWaitNeoUnlockVerify:
		txid = event.NeoUnlockTxhash
	default:
		return CchTxVerifyStatUnknown, errors.New("Nep5VerifyByTxid:bad event status")
	}
	if len(txid) < WrapperTxHashMinLen {
		t.logger.Error("Nep5VerifyByTxid bad txid:%s", txid)
		return CchTxVerifyStatUnknown, errors.New("Nep5VerifyByTxid:bad txid")
	}
	time.Sleep(3 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	for i := 0; i < 50; i++ {
		<-ticker.C
		txinfo, err := t.Nep5GetTxInfo(txid)
		if err == nil {
			switch event.Status {
			case cchNep5MortgageStatusWaitNeoLockVerify:
			case cchNep5MortgageStatusWaitNeoUnlockVerify:
			case cchEthRedemptionStatusWaitNeoLockVerify:
			case cchEthRedemptionStatusWaitNeoUnlockVerify:
				if txinfo.Amount != event.Amount {
					t.logger.Debugf("Nep5VerifyByTxid: amount(%d %d) check err", txinfo.Amount, event.Amount)
					return CchTxVerifyStatFailed, errors.New("amount check err")
				}
			}
			return CchTxVerifyStatOk, nil
		}
	}
	return CchTxVerifyStatUnknown, errors.New("get application log time out")
}

func (t *Transaction) BuildContractVerifyScript(operation string, args []interface{}) (*neotransaction.Script, error) {
	script := &neotransaction.Script{}
	// 压栈脚本为空

	// 创建鉴权脚本，填充合约参数
	opscript, err := t.CreateScript(operation, args, false)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}
	script.VerificationScript = make([]byte, len(opscript))
	copy(script.VerificationScript, opscript)
	script.VrifScriptLength.Value = uint64(len(script.VerificationScript))
	return script, nil
}

//NeoBlockNumberSysn
func (w *WrapperServer) NeoBlockNumberSysn() {
	curnum, err := w.sc.WsqlLastBlockNumGet(CchBlockTypeNeo)
	if err != nil {
		w.logger.Error("WsqlLastBlockNumGet err:", err)
		return
	}
	gWrapperStats.LastNeoBlocknum = curnum
	gWrapperStats.CurrentNeoBlocknum = curnum
}

//EthUpdateBlockNumber 定时任务，同步当前区块高度
func (w *WrapperServer) NeoUpdateBlockNumber() {
	//定时查询最新块高度
	d := time.Duration(time.Second * 10)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		<-t.C
		header, err := neocliapi.FetchBlockHeight(neoEndPoint)
		if err != nil {
			w.logger.Error("NeoUpdateBlockNumber err:", err)
			continue
		}
		if int64(header) != gWrapperStats.CurrentNeoBlocknum {
			gWrapperStats.CurrentNeoBlocknum = int64(header)
			w.sc.WsqlBlockNumberUpdateLogInsert(CchBlockTypeNeo, gWrapperStats.CurrentNeoBlocknum, "update neo blocknum")
		}
	}
}

//ParseNep5TransfersResult
func (t *Transaction) ParseNep5TransfersResult(i interface{}) ([]*DBNeoEventLogTBL, error) {
	//trans := make([]*DBNeoEventLogTBL, 0)
	r, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transaction data error: %s", i)
	}
	res, ok := r["result"]
	if !ok {
		return nil, errors.New("transaction data has no result")
	}
	es, ok := res.([]interface{})
	if !ok || len(es) < 1 {
		return nil, errors.New("executions data not found")
	}
	e, ok := es[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("executions data error")
	}
	sent, ok := e["sent"]
	if ok {
		vs, ok := sent.(map[string]interface{})
		if ok {
			t.logger.Debugf("get sent(%v) vs(%v)", sent, vs)
		}

	}
	return nil, nil
}

//NeoGetNep5Transfers
func (t *Transaction) NeoGetNep5Transfers(starttime, endtime int64, addr string) ([]*DBNeoEventLogTBL, error) {
	var param []interface{}
	if starttime != 0 {
		param = []interface{}{addr, starttime, endtime}
	} else {
		param = []interface{}{addr}
	}

	i, err := request.HttpRequest("getnep5transfers", param, neoEndPoint)
	if err != nil {
		return nil, err
	}
	return t.ParseNep5TransfersResult(i)
}

//NeoGetNep5Transfers
func (t *Transaction) NeoGetNep5EventListen() {
	endtime := time.Now().Unix()
	starttime := endtime - 100000
	lasttime := endtime
	t.NeoGetNep5Transfers(starttime, endtime, WrapperNeoContract)
	//定时查询最新块高度
	//d := time.Duration(time.Second * 5)
	timer := time.NewTicker(time.Second * 5)
	defer timer.Stop()
	for {
		<-timer.C
		starttime = lasttime
		endtime = time.Now().Unix()
		lasttime = endtime
		t.NeoGetNep5Transfers(starttime, endtime, WrapperNeoContract)
	}
}
