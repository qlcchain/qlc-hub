package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qlcchain/qlc-hub/config"
	"github.com/qlcchain/qlc-hub/pkg/eth"
	"github.com/qlcchain/qlc-hub/pkg/log"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var averageGas int64

func GetBestGas(ctx context.Context, url string) {
	logger := log.NewLogger("api/gas")

	var err error
	averageGas, err = getAverageGas(url)
	if err != nil {
		logger.Error(err)
	}
	fmt.Println(averageGas)
	vTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-vTicker.C:
			if averageGas, err = getAverageGas(url); err != nil {
				logger.Error(err)
			}
		}
	}
}

func getAverageGas(url string) (int64, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode > 200 {
		return 0, fmt.Errorf("response status: %d", response.StatusCode)
	}

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(bs, &ret)
	if err != nil {
		return 0, err
	}

	if ave, ok := ret["average"]; ok {
		if average, ok := ave.(float64); ok {
			return int64(average), nil
		} else {
			return 0, errors.New("invalid type")
		}
	} else {
		return 0, errors.New("average not found")
	}
}

type gasManageDaily struct {
	requestTimes     int64
	initialTime      int64
	gasUsed          int64
	gasEndpointIndex int
}

var gasManage = &gasManageDaily{}
var mLock = sync.Mutex{}

func checkGas(cfg *config.Config, e *eth.Transaction) error {
	balance, err := e.Balance(cfg.EthereumCfg.OwnerAddress)
	if err != nil {
		return err
	}
	if balance < 1 {
		return fmt.Errorf("owner balance %d is not sufficient", balance)
	}

	if gasManage.gasUsed > cfg.EthereumCfg.MaxGasPerDay {
		return fmt.Errorf("gas limit has already exhausted, used %d", gasManage.gasUsed)
	}

	if gasManage.requestTimes >= cfg.EthereumCfg.MaxRequestPerDay {
		gasManage.gasEndpointIndex++
		if len(cfg.EthereumCfg.EndPoint) > gasManage.gasEndpointIndex {
			eClient, err := ethclient.Dial(cfg.EthereumCfg.EndPoint[gasManage.gasEndpointIndex])
			if err != nil {
				return fmt.Errorf("eth dail: %s", err)
			}
			e.SetClient(eClient)
			gasManage.requestTimes = 0
		}
	}
	return nil
}

func updateGas(txHash string, e *eth.Transaction, logger *zap.SugaredLogger) {
	mLock.Lock()
	defer mLock.Unlock()

	gasManage.requestTimes++
	tx, _, err := e.Client().TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		logger.Error(err)
	}
	gasCost := tx.Cost()
	gasManage.gasUsed = gasManage.gasUsed + gasCost.Int64()
}

func resetGas(ctx context.Context) {
	vTicker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			return
		case <-vTicker.C:
			mLock.Lock()
			gasManage.gasUsed = 0
			gasManage.requestTimes = 0
			mLock.Unlock()
		}
	}
}
