package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/types"
)

func waitForLockerState(rHash string, lockerState types.LockerState) {
	cTicker := time.NewTicker(6 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getHashTimerState(rHash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("====== ", state)
		if state == types.LockerStateToString(lockerState) {
			return
		}
	}
	logger.Fatal("timeout")
}

func getHashTimerState(rHash string) (string, error) {
	bs, err := get(fmt.Sprintf("%s/debug/lockerState?value=%s", hubUrl, rHash))
	if err != nil {
		return "", err
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(bs, &ret)
	if err != nil {
		return "", err
	}
	return ret["stateStr"].(string), nil
}

func get(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s", url), nil)
	if err != nil {
		return nil, fmt.Errorf("request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("client do: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 200 {
		return nil, fmt.Errorf("StatusCode : %d", response.StatusCode)
	}

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll : %s", err)
	}
	return bs, nil
}

func post(paras string, url string) (bool, error) {
	jsonStr := []byte(paras)
	ioBody := bytes.NewBuffer(jsonStr)
	request, err := http.NewRequest("POST", url, ioBody)
	if err != nil {
		logger.Fatal("request ", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logger.Fatal("do ", err)
	}
	defer response.Body.Close()
	if response.StatusCode > 200 {
		logger.Fatalf("%d status code returned ", response.StatusCode)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Fatal(err)
	}

	ret := make(map[string]interface{})
	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		logger.Fatal(err)
	}
	if r, ok := ret["value"]; ok != false {
		return r.(bool), nil
	} else {
		if e, ok := ret["error"]; ok != false {
			return false, fmt.Errorf("%s", e)
		}
		return false, fmt.Errorf("response has no result")
	}
}

func getContractAddress() (string, string) {
	bs, err := get(fmt.Sprintf("%s/debug/ping", hubUrl))
	if err != nil {
		logger.Fatal(err)
	}
	ret := make(map[string]string)
	err = json.Unmarshal(bs, &ret)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(ret)
	return ret["ethContract"], ret["neoContract"]
}
