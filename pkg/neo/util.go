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
	return request.Param{
		Type: request.FuncParamT,
		Value: request.FuncParam{
			Type: smartcontract.ByteArrayType,
			Value: request.Param{
				Type:  request.ArrayT,
				Value: addrHex,
			},
		},
	}
}

func HashParam(hash string) request.Param {
	return request.Param{
		Type: request.FuncParamT,
		Value: request.FuncParam{
			Type: smartcontract.ByteArrayType,
			Value: request.Param{
				Type:  request.StringT,
				Value: hash,
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

func ArrayTypeParams(params []request.Param) request.Param {
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
