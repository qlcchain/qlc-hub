package wrapper

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	_ "fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	_ "strings"
	"time"

	"github.com/ethereum/go-ethereum"
	_ "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	QLCChain "github.com/qlcchain/qlc-hub/contract"
	"github.com/shopspring/decimal"
)

//测试网地址
const EthRinkebyHttpsLink string = "https://rinkeby.infura.io/v3/a63d9065622f422588b94a6b52f71e5a"
const EthRinkebyWssLink string = "wss://rinkeby.infura.io/ws/v3/a63d9065622f422588b94a6b52f71e5a"

const EthGetHashTimerLoopTime = 5 * time.Second

const (
	EthEventStatusIssueLock     int64 = 0 //issueLock
	EthEventStatusIssueUnLock         = 1 //issueUnlock
	EthEventStatusIssueFetch          = 2 //issueFetch
	EthEventStatusDestoryLock         = 3 //destoryLock
	EthEventStatusDestoryUnlock       = 4 //destoryUnlock
	EthEventStatusDestoryFetch        = 5 //destoryFetch
)

// IsValidAddress validate hex address
func (w *WrapperServer) IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func (w *WrapperServer) IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal wei to decimals
func (w *WrapperServer) ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

// ToWei decimals to wei
func (w *WrapperServer) ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

// CalcGasCost calculate gas cost given gas limit (units) and gas price (wei)
func (w *WrapperServer) CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

// SigRSV signatures R S V returned as arrays
func (w *WrapperServer) SigRSV(isig interface{}) ([32]byte, [32]byte, uint8) {
	var sig []byte
	switch v := isig.(type) {
	case []byte:
		sig = v
	case string:
		sig, _ = hexutil.Decode(v)
	}

	sigstr := common.Bytes2Hex(sig)
	rS := sigstr[0:64]
	sS := sigstr[64:128]
	R := [32]byte{}
	S := [32]byte{}
	copy(R[:], common.FromHex(rS))
	copy(S[:], common.FromHex(sS))
	vStr := sigstr[128:130]
	vI, _ := strconv.Atoi(vStr)
	V := uint8(vI + 27)

	return R, S, V
}

//WrapperEthClientConnect
func (w *WrapperServer) WrapperEthClientConnect() (c *ethclient.Client, err error) {
	for i := 0; i < 100; i++ {
		client, err := ethclient.Dial(EthRinkebyWssLink)
		if err != nil {
			w.logger.Error(err)
			time.Sleep(time.Second * 3)
		}
		return client, nil
	}
	return nil, err
}

//WrapperEthListen eth listen
func (w *WrapperServer) WrapperEthListen() {
	var logindex int
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		time.Sleep(5)
	}
	contractAddress := common.HexToAddress(WrapperEthContract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	filogger, err := QLCChain.NewQLCChainFilterer(contractAddress, client)
	if err != nil {
		w.logger.Error("NewQLCChainFilterer err:", err)
		return
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		w.logger.Error("SubscribeFilterLogs err:", err)
		return
	}
	logindex = 0
	for {
		select {
		case err := <-sub.Err():
			w.logger.Error("logs err:", err)
		case vLog := <-logs:
			event, err := filogger.ParseLockedState(vLog)
			if err != nil {
				w.logger.Debugf("unknown vlog")
				break
			}
			w.logger.Debugf("get logs: block(%s),blocknum(%d),txhash(%s)", vLog.BlockHash.Hex(), vLog.BlockNumber, vLog.TxHash.Hex())
			//w.logger.Debugf("get RHash:",event.RHash)
			//w.logger.Debugf("get State:",event.State)
			rhash := hex.EncodeToString(event.RHash[:])
			action := event.State.Int64()
			txhash := vLog.TxHash.Hex()
			w.logger.Debugf("get log%d action :%d, rash %s", logindex, action, rhash)
			logindex++
			go w.EthGetHashTimerLoop(action, rhash, txhash)
		}
	}
}

//EthGetBlockByTxhash get block by txhash
func (w *WrapperServer) EthGetBlockByTxhash(infotype int64, txhashstring string) (result int64, info string, err error) {
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, "", errors.New("eth conn failed")
	}
	txHash := common.HexToHash(txhashstring)
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadTxHash, "", errors.New("eth get by txhash failed")
	}
	var txinfo string
	txinfo = txinfo + tx.To().Hex() + "_"
	if infotype == CchEthGetTransTypeGetAll {
		txinfo = txinfo + tx.Value().String()
		if isPending == false {
			txinfo = txinfo + "_pending_false"
		} else {
			txinfo = txinfo + "_pending_true"
		}
	}
	return CchEthIssueRetOK, txinfo, nil
}

//EthGetAccountByAddr get account by address
func (w *WrapperServer) EthGetAccountByAddr(addr string) (result int64, info string, err error) {
	var accountinfo string
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, "", errors.New("eth conn failed")
	}
	account := common.HexToAddress(addr)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, "", errors.New("eth BalanceAt failed")
	}
	w.logger.Debug("balance:", balance) // 25893180161173005034
	accountinfo = "balance" + balance.String()
	return CchEthIssueRetOK, accountinfo, nil
}

//EthContractPredeal predeal before eth smartcontract call
func (w *WrapperServer) EthContractPredeal() (ret int64, qin *QLCChain.QLCChainTransactor, opts *bind.TransactOpts, err error) {
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, nil, nil, errors.New("eth conn failed")
	}

	privateKey, err := crypto.HexToECDSA(WrapperEthPrikey)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth prikey HexToECDSA failed")
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		w.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth prikey get publicKey failed")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth PendingNonceAt failed")
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, nil, nil, errors.New("eth SuggestGasPrice failed")
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(1000000) // in units
	auth.GasPrice = gasPrice
	address := common.HexToAddress(WrapperEthContract)
	instance, err := QLCChain.NewQLCChainTransactor(address, client)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth NewQLCChainTransactor failed")
	}
	return CchEthIssueRetOK, instance, auth, nil
}

//EthContractIssueLock issuelock
func (w *WrapperServer) EthContractIssueLock(amount int64, lockhash string) (result int64, txhash string, err error) {
	ret, instance, opts, err := w.EthContractPredeal()
	if err != nil {
		return ret, "", err
	}
	bigAmount := big.NewInt(amount * WrapperGasWeiNum)
	var lockarray [32]byte
	lock, err := hex.DecodeString(lockhash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadLockHash, "", err
	}
	copy(lockarray[:], lock)
	// copy(lockarray[:],[]byte(lockhash))
	//w.logger.Debugf("get bigAmount % lockarray %",bigAmount,lockarray)
	tx, err := instance.IssueLock(opts, lockarray, bigAmount)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, "", errors.New("eth call IssueLock failed")
	}
	w.logger.Debugf("EthContractIssueLock: lock hash %s amount %d", lockhash, amount)
	return CchEthIssueRetOK, tx.Hash().Hex(), nil
}

//EthContractIssueFetch issueunFetch
func (w *WrapperServer) EthContractIssueFetch(lockhash string) (result int64, txhash string, err error) {
	ret, instance, opts, err := w.EthContractPredeal()
	if err != nil {
		return ret, "", err
	}
	var lockarray [32]byte
	lock, err := hex.DecodeString(lockhash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadLockHash, "", err
	}
	copy(lockarray[:], lock)
	// copy(lockarray[:],[]byte(lockhash))
	tx, err := instance.IssueFetch(opts, lockarray)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, "", errors.New("eth call IssueFetch failed")
	}
	return CchEthIssueRetOK, tx.Hash().Hex(), nil
}

//EthContractDestoryUnlock destory unlock
func (w *WrapperServer) EthContractDestoryUnlock(lockhash string, locksource string) (result int64, txhash string, err error) {
	var lockarray [32]byte
	var sourcearray [32]byte
	ret, instance, opts, err := w.EthContractPredeal()
	if err != nil {
		return ret, "", err
	}
	w.logger.Debugf("DestoryUnlock get lockhash %s locksource %s", lockhash, locksource)
	lock, err := hex.DecodeString(lockhash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadLockHash, "", err
	}
	copy(lockarray[:], lock)
	//copy(lockarray[:],[]byte(lockhash))
	w.logger.Debugf("DestoryUnlock get lockarray", lockarray)
	// source,err := hex.DecodeString(locksource)
	// if err != nil {
	// 	w.logger.Error(err)
	// 	return CchEthIssueRetBadLockHash,"",err
	// }
	// copy(sourcearray[:],source)
	copy(sourcearray[:], []byte(locksource))
	w.logger.Debugf("DestoryUnlock get sourcearray", sourcearray)
	tx, err := instance.DestoryUnlock(opts, lockarray, sourcearray)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, "", errors.New("eth call DestoryUnlock failed")
	}
	return CchEthIssueRetOK, tx.Hash().Hex(), nil
}

//EthContractUserCallerPredeal predeal before eth smartcontract call
func (w *WrapperServer) EthContractUserCallerPredeal(ucallerkey string) (ret int64, qin *QLCChain.QLCChainTransactor, opts *bind.TransactOpts, err error) {
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, nil, nil, errors.New("eth conn failed")
	}

	privateKey, err := crypto.HexToECDSA(ucallerkey)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth prikey HexToECDSA failed")
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		w.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth prikey get publicKey failed")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth PendingNonceAt failed")
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, nil, nil, errors.New("eth SuggestGasPrice failed")
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(1000000) // in units
	auth.GasPrice = gasPrice
	address := common.HexToAddress(WrapperEthContract)
	//w.logger.Debugf("get nonce: %,GasPrice %",nonce,auth.GasPrice)
	instance, err := QLCChain.NewQLCChainTransactor(address, client)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, nil, nil, errors.New("eth NewQLCChainTransactor failed")
	}
	return CchEthIssueRetOK, instance, auth, nil
}

//EthContractUcallerDestoryLock user call destory unlock
func (w *WrapperServer) EthContractUcallerDestoryLock(amount int64, lockhash string) (result int64, txhash string, err error) {
	ret, instance, opts, err := w.EthContractUserCallerPredeal(WrapperEthUserPrikey)
	if err != nil {
		return ret, "", err
	}
	bigAmount := big.NewInt(amount * WrapperGasWeiNum)
	var lockarray [32]byte
	lock, err := hex.DecodeString(lockhash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadLockHash, "", err
	}
	copy(lockarray[:], lock)
	// copy(lockarray[:],[]byte(lockhash))
	w.logger.Debugf("EthContractUcallerDestoryLock get bigAmount % lockarray %", bigAmount, lockarray)
	owneraddress := common.HexToAddress(WrapperEthAccount)
	tx, err := instance.DestoryLock(opts, lockarray, bigAmount, owneraddress)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, "", errors.New("eth call DestoryLock failed")
	}
	return CchEthIssueRetOK, tx.Hash().Hex(), nil
}

//EthGetHashTimer user call destory unlock
func (w *WrapperServer) EthGetHashTimer(lockhash string) (result, amount, locknum, unlocknum int64, account, locksource string, err error) {
	var lockarray [32]byte
	var callops = bind.CallOpts{}
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, 0, 0, 0, "", "", errors.New("eth conn failed")
	}
	contractaddress := common.HexToAddress(WrapperEthContract)
	instance, err := QLCChain.NewQLCChainCaller(contractaddress, client)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetClientConnFailed, 0, 0, 0, "", "", errors.New("eth get instance failed")
	}
	lock, err := hex.DecodeString(lockhash)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadLockHash, 0, 0, 0, "", "", err
	}
	copy(lockarray[:], lock)
	source, amountbig, addr, lockno, unlockno, _, err := instance.HashTimer(&callops, lockarray)
	if err != nil {
		w.logger.Error(err)
		return CchEthIssueRetBadKey, 0, 0, 0, "", "", errors.New("eth call HashTimer failed")
	}
	ramount := amountbig.Int64()
	rlocknum := lockno.Int64()
	runlocknum := unlockno.Int64()
	//w.logger.Debugf("HashTimer get source:",source)
	//w.logger.Debugf("HashTimer get addr:",addr)
	rsource := string(source[:])
	raddr := hex.EncodeToString(addr[:])
	w.logger.Debugf("HashTimer get amount:%d,locknum:%d,unlocknum:%d", ramount, rlocknum, runlocknum)
	return CchEthIssueRetOK, ramount, rlocknum, runlocknum, raddr, rsource, nil
}

func (w *WrapperServer) EthVerifyByLockhash(event *EventInfo) {
	initstatus := event.Status
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 100; i++ {
		<-ticker.C
		//这里是为了避免状态已经被其他事件触发改变
		if initstatus != event.Status {
			w.logger.Debugf("event(%s) break EthVerifyByLockhash")
			return
		}
		_, _, locknum, unlocknum, account, locksource, err := w.EthGetHashTimer(event.LockHash)
		if err == nil {
			switch event.Status {
			case cchNep5MortgageStatusWaitEthLockVerify:
				//记录locknum
				event.EthLockNum = locknum
				event.Status = cchNep5MortgageStatusWaitClaim
			case cchNep5MortgageStatusWaitEthUnlockVerify:
				event.UnlockNum = unlocknum
				event.UserAccount = account
				event.HashSource = locksource
				event.Status = cchNep5MortgageStatusTryNeoUnlock
			case cchNep5MortgageStatusTimeoutDestroyVerify:
				event.UnlockNum = unlocknum
				event.Status = cchNep5MortgageStatusTimeoutDestroyOk
			default:
				break
			}
			w.logger.Debugf("EthVerifyByLockhash:status in(%d) out(%d)", initstatus, event.Status)
			if event.Status != initstatus {
				w.sc.DbEventUpdate(event)
				go w.eventStatusUpdateMsgPush(event, event.Status)
			}
			return
		}
	}
	//timeout
	switch event.Status {
	case cchNep5MortgageStatusWaitEthLockVerify:
	case cchNep5MortgageStatusWaitEthUnlockVerify:
	case cchNep5MortgageStatusTimeoutDestroyVerify:
		event.Status = cchNep5MortgageStatusFailed
		event.Errno = CchEventRunErrMortgageNep5VerifyFailed
		err := w.WrapperEventUpdateStatByLockhash(event.Type, event.Status, event.Errno, event.LockHash)
		if err != nil {
			w.logger.Error("WrapperEventUpdateStatByLockhash: err", err)
		}
	default:
		break
	}
}

//EthGetHashTimerDeal
func (w *WrapperServer) EthGetHashTimerDeal(action, amount, locknum, unlocknum int64, lockhash, locksource, txhash, addr string) {
	var newstat int64
	if action == EthEventStatusDestoryLock {
		event := RedemptionEvent[lockhash]
		if event != nil {
			newstat = cchEthRedemptionStatusTimeoutUnlockOk
			err := w.WrapperEventUpdateStatByLockhash(cchEventTypeRedemption, newstat, 0, lockhash)
			if err != nil {
				w.logger.Error("WrapperEventUpdateStatByLockhash err:", err)
			}
		} else {
			newstat = cchEthRedemptionStatusTryNeoLock
			err := w.WrapperEventInsert(newstat, amount, cchEventTypeRedemption, locknum, lockhash, txhash, addr)
			if err != nil {
				w.logger.Error("WrapperEventInsert err:", err)
			}
		}
	} else if action == EthEventStatusDestoryFetch {
		newstat = cchEthRedemptionStatusTimeoutUnlockOk
		err := w.WrapperEventUpdateStatByLockhash(cchEventTypeRedemption, newstat, 0, lockhash)
		if err != nil {
			w.logger.Error("WrapperEventUpdateStatByLockhash err:", err)
		}
	} else if action == EthEventStatusIssueLock {
		newstat = cchNep5MortgageStatusWaitClaim
		err := w.WrapperEventUpdateStatByLockhash(cchEventTypeMortgage, newstat, 0, lockhash)
		if err != nil {
			w.logger.Error("WrapperEventUpdateStatByLockhash err:", err)
		}
	}
	w.logger.Debugf("EthGetHashTimerDeal:action %d,lockhash %s stat %d", action, lockhash, newstat)
}

//EthGetHashTimerLoop
func (w *WrapperServer) EthGetHashTimerLoop(action int64, lockhash, txhash string) {
	ticker := time.NewTicker(EthGetHashTimerLoopTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//GetHashTimer
			ret, amount, locknum, unlocknum, account, locksource, err := w.EthGetHashTimer(lockhash)
			if err == nil && ret == CchEthIssueRetOK {
				//w.logger.Debugf("EthGetHashTimer:get amount:%,locknum:%",amount,locknum)
				if locknum > 0 {
					w.logger.Debugf("EthGetHashTimerLoop out,lockhash:(%s)", lockhash)
					w.EthGetHashTimerDeal(action, amount, locknum, unlocknum, lockhash, locksource, txhash, account)
					return
				}
			}
		}
	}
}

//EthBlockNumbersysn
func (w *WrapperServer) EthBlockNumbersysn() {
	curnum, err := w.sc.WsqlLastBlockNumGet(CchBlockTypeEth)
	if err != nil {
		w.logger.Error("WsqlLastBlockNumGet err:", err)
		return
	}
	gWrapperStats.LastEthBlocknum = curnum
	gWrapperStats.CurrentEthBlocknum = curnum
}

//EthUpdateBlockNumber 定时任务，同步当前区块高度
func (w *WrapperServer) EthUpdateBlockNumber() {
	client, err := w.WrapperEthClientConnect()
	if err != nil {
		w.logger.Error("EthUpdateBlockNumber err:", err)
		return
	}
	//定时查询最新块高度
	d := time.Duration(time.Second * 10)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		<-t.C
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			w.logger.Error("EthUpdateBlockNumber err:", err)
			continue
		}
		if header.Number.Int64() != gWrapperStats.CurrentEthBlocknum {
			gWrapperStats.CurrentEthBlocknum = header.Number.Int64()
			w.sc.WsqlBlockNumberUpdateLogInsert(CchBlockTypeEth, gWrapperStats.CurrentEthBlocknum, "update eth blocknum")
		}
	}
}
