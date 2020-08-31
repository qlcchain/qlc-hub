package neo

import (
	"encoding/hex"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpc/request"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
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

func IntegerTypeParam(v int) request.Param { //todo int64
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
