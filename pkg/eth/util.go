package eth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
)

func (t *Transaction) GetBestGas() (*big.Int, error) {
	suggestPrice, err := t.client.SuggestGasPrice(context.Background()) // unit is wei
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %s", err)
	}
	if t.averageGas == 0 {
		return suggestPrice, nil
	}
	if suggestPrice.Int64() < t.averageGas {
		return suggestPrice, nil
	} else {
		return big.NewInt(t.averageGas), nil
	}
}

func (t *Transaction) Gas() (int64, int64, error) {
	suggestPrice, err := t.client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, 0, fmt.Errorf("suggest gas price: %s", err)
	}
	return t.averageGas, suggestPrice.Int64(), err
}

func (t *Transaction) updateAverageGas() {
	averageGas, err := getAverageGas(t.gasUrl)
	if err != nil {
		t.logger.Error(err)
	}
	t.averageGas = averageGas
	vTicker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-vTicker.C:
			if averageGas, err := getAverageGas(t.gasUrl); err != nil {
				t.logger.Error(err)
			} else {
				t.averageGas = averageGas
				t.logger.Debugf("update average gas: %d", averageGas)
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
			return int64(average) * 1e8, nil //average unit is x10 gwei， convert to wei
		} else {
			return 0, errors.New("invalid type")
		}
	} else {
		return 0, errors.New("average not found")
	}
}
