package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/qlcchain/qlc-hub/log"
)

func HttpRequest(method string, bodyParameters []interface{}, nodeURI string) (interface{}, error) {
	log := log.NewLogger("HttpRequest")
	var body []byte
	var err error

	if bodyParameters == nil {
		body, err = NewBody(method)
		if err != nil {
			return nil, fmt.Errorf("NewBody error: %s ", err)
		}
	} else {
		body, err = NewBodyWithParameters(method, bodyParameters)
		if err != nil {
			return nil, fmt.Errorf("NewBodyWithParameters error: %s ", err)
		}
	}
	log.Debug("request: ", string(body))

	ioBody := bytes.NewReader(body)

	request, err := http.NewRequest("POST", nodeURI, ioBody)
	if err != nil {
		return nil, fmt.Errorf("NewRequest error: %s ", err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("sends an HTTP request error: %s ", err)
	}
	defer response.Body.Close()

	if response.StatusCode > 200 {
		return nil, fmt.Errorf("%d status code returned ", response.StatusCode)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]interface{})
	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("executeRequest error: %s ", err)
	}
	log.Debug("response: ", string(bytes))

	if r, ok := ret["result"]; ok != false {
		return r, nil
	} else {
		if e, ok := ret["error"]; ok != false {
			return nil, fmt.Errorf("%s", e)
		}
		return nil, fmt.Errorf("response has no result")
	}
}
