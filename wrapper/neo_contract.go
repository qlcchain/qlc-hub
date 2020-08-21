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
	contractHash = "30f69798a129527b4996d6dd8e974cc15d51403d"
	neoEndPoint  = "http://seed1.ngd.network:20332"
	//wrapperEndPoint = "https://nep5-test.qlcchain.online"
	//wrapperEndPoint = "http://47.103.54.171:19740"
)

var (
	OpLock              = "lock"
	OpUnLock            = "unlock"
	OpWrapperLock       = "wrapperLock"
	OpWrapperUnlock     = "wrapperUnlock"
	OpRefundWrapper     = "refundWrapper"
	OpGetLockInfo       = "getLockInfo"
	OpGetApplicationLog = "getapplicationlog"
)

type Transaction struct {
	neoNode      string
	contract     string
	officialNode []string
	logger       *zap.SugaredLogger
}

func NewTransaction(neoNode string, contract string, officialNode []string) *Transaction {
	return &Transaction{
		neoNode:      neoNode,
		contract:     contract,
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

	t.logger.Debug("txID: ", tx.TXID())
	rawtx := tx.RawTransactionString()
	b := neocliapi.SendRawTransaction(t.neoNode, rawtx)
	if !b {
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
	ticker := time.NewTicker(1 * time.Second)
	for i := 0; i < 300; i++ {
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

func (t *Transaction) Nep5WrapperLock(amount, locknum int64, addr, lockhash string) (string, error) {
	params := lockhash + addr + strconv.FormatInt(amount, 10) + strconv.FormatInt(locknum, 10)
	b, err := hex.DecodeString(params)
	if err != nil {
		return "", err
	}
	args := []interface{}{neoutils.Reverse(b)}
	script, err := t.CreateScript(OpWrapperLock, args, true)
	if err != nil {
		return "", err
	}
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5WrapperUnlock(locksource, addr string) (string, error) {
	params := locksource + addr
	b, err := hex.DecodeString(params)
	if err != nil {
		return "", err
	}
	args := []interface{}{neoutils.Reverse(b)}
	script, err := t.CreateScript(OpWrapperUnlock, args, true)
	if err != nil {
		return "", err
	}
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5WrapperRefund(locksource string) (string, error) {
	params := locksource
	b, err := hex.DecodeString(params)
	if err != nil {
		return "", err
	}
	args := []interface{}{neoutils.Reverse(b)}
	script, err := t.CreateScript(OpRefundWrapper, args, true)
	if err != nil {
		return "", err
	}
	id, err := t.CreateTransaction(script, WrapperNeoPrikey)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (t *Transaction) Nep5TransactionVerify(txid string) (status int, err error) {
	_, err = t.WaitApplicationLog(txid)
	if err != nil {
		return CchTransactionVerifyStatusFalse, err
	}
	return CchTransactionVerifyStatusTrue, nil
}
