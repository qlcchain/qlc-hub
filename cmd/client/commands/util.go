package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/types"
)

func hubWaitingForWithdrawEthTimeout(rHash string) bool {
	cTicker := time.NewTicker(40 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getLockerState(rHash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Printf("rHash [%s] state is [%s] \n", rHash, state["stateStr"])
		if state["ethTimeout"].(bool) {
			return true
		}
	}
	log.Fatal("timeout")
	return false
}

func hubWaitingForDepositNeoTimeout(rHash string) bool {
	cTicker := time.NewTicker(40 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getLockerState(rHash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Printf("rHash [%s] state is [%s] \n", rHash, state["stateStr"])
		if state["neoTimeout"].(bool) {
			return true
		}
	}
	log.Fatal("timeout")
	return false
}

func hubWaitingForLockerState(rHash string, lockerState types.LockerState) bool {
	cTicker := time.NewTicker(30 * time.Second)
	for i := 0; i < 100; i++ {
		<-cTicker.C
		state, err := getLockerState(rHash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		log.Printf("rHash [%s] state is [%s] \n", rHash, state["stateStr"])
		if state["fail"].(bool) {
			log.Printf("rHash [%s] fail: [%s] \n", rHash, state["remark"].(string))
			return false
		}
		if state["stateStr"].(string) == types.LockerStateToString(lockerState) {
			return true
		}
	}
	log.Fatal("timeout")
	return false
}

func getLockerState(rHash string) (map[string]interface{}, error) {
	ret, err := get(fmt.Sprintf("%s/info/lockerInfo?value=%s", hubUrl, rHash))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func getPing(addr string) (map[string]interface{}, error) {
	ret, err := get(fmt.Sprintf("%s/info/ping?value=%s", hubUrl, addr))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func get(url string) (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s", url), nil)
	if err != nil {
		return nil, fmt.Errorf("request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("authorization", hubCmd.HubToken)

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
	ret := make(map[string]interface{})
	err = json.Unmarshal(bs, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func post(paras string, url string) (interface{}, error) {
	jsonStr := []byte(paras)
	ioBody := bytes.NewBuffer(jsonStr)
	request, err := http.NewRequest("POST", url, ioBody)
	if err != nil {
		log.Fatal("request ", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("authorization", hubCmd.HubToken)
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("do ", err)
	}
	defer response.Body.Close()
	if response.StatusCode > 200 {
		log.Fatalf("%d status code returned ,%s", response.StatusCode, url)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	ret := make(map[string]interface{})
	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		log.Fatal(err)
	}
	if r, ok := ret["value"]; ok != false {
		return r, nil
	} else {
		if e, ok := ret["error"]; ok != false {
			return false, fmt.Errorf("%s", e)
		}
		return nil, fmt.Errorf("response has no result")
	}
}
