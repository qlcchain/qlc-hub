package neo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
)

func AddressParam(addr string) request.Param {
	addrUint, _ := address.StringToUint160(addr)
	addrHex := hex.EncodeToString(addrUint.BytesBE())
	return ArrayTypeParam(addrHex)
}

func ArrayTypeParam(hexStr string) request.Param {
	return request.Param{
		Type: request.FuncParamT,
		Value: request.FuncParam{
			Type: smartcontract.ByteArrayType,
			Value: request.Param{
				Type:  request.ArrayT,
				Value: hexStr,
			},
		},
	}
}

func IntegerTypeParam(v int) request.Param {
	return request.Param{
		Type: request.FuncParamT,
		Value: request.FuncParam{
			Type: smartcontract.IntegerType,
			Value: request.Param{
				Type:  request.NumberT,
				Value: v,
			},
		},
	}
}

func StringTypeParam(v string) request.Param {
	return request.Param{
		Type: request.FuncParamT,
		Value: request.FuncParam{
			Type: smartcontract.StringType,
			Value: request.Param{
				Type:  request.StringT,
				Value: v,
			},
		},
	}
}

func ArrayParams(params []request.Param) request.Param {
	return request.Param{
		Type:  request.ArrayT,
		Value: params,
	}
}

func FunctionName(name string) request.Param {
	return request.Param{
		Type:  request.StringT,
		Value: name,
	}
}

func StackToSwapInfo(stack []smartcontract.Parameter) (map[string]interface{}, error) {
	value := stack[0].Value
	if v, ok := value.([]byte); ok {
		return nil, fmt.Errorf("invalid value type: %s", string(v))
	} else if data, ok := value.([]smartcontract.Parameter); ok {
		if len(data) != 8 {
			return nil, fmt.Errorf("invalid data, exp: 15, got: %d", len(data))
		}
		result := make(map[string]interface{}, 0)
		for idx, v := range data {
			k := fields[idx]
			switch v.Type {
			case smartcontract.BoolType:
				continue
			case smartcontract.ByteArrayType:
				d, ok := v.Value.([]byte)
				if !ok {
					return nil, errors.New("invalid ByteArray item")
				}
				result[k] = bytesTo(k, d)
			case smartcontract.IntegerType:
				i, ok := v.Value.(int64)
				if !ok {
					return nil, errors.New("invalid Integer item")
				}
				result[k] = intTo(k, i)
			}
		}

		return result, nil
	} else {
		return nil, errors.New("invalid data")
	}
}

func bytesTo(key string, v []byte) interface{} {
	if t, ok := types[key]; ok {
		switch t {
		case "amount":
			d2 := util.ArrayReverse(v)
			return big.NewInt(0).SetBytes(d2)
		case "int":
			return emit.BytesToInt(v).Int64()
		case "neo":
			a, _ := util.Uint160DecodeBytesBE(v)
			return address.Uint160ToString(a)
		case "eth":
			return common.BytesToAddress(v)
		case "hash":
			return hex.EncodeToString(util.ArrayReverse(v))
		case "text":
			return string(v)
		}
	}
	return hex.EncodeToString(v)
}

func intTo(key string, i int64) interface{} {
	if v, ok := types[key]; ok {
		if v == "time" {
			return time.Unix(i, 0)
		}
	}
	return i
}

var (
	types = map[string]string{
		"fromAddress":    "neo",
		"toAddress":      "neo",
		"txid":           "hash",
		"userEthAddress": "eth",
		"amount":         "amount",
		"timestamp":      "amount",
		"blockHeight":    "amount",
		"type":           "int",
	}
	fields = []string{
		"fromAddress",
		"toAddress",
		"txid",
		"userEthAddress",
		"amount",
		"timestamp",
		"blockHeight",
		"type",
	}
)
