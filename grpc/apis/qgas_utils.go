package apis

import (
	"math/big"

	"github.com/qlcchain/qlc-go-sdk/pkg/types"

	pb "github.com/qlcchain/qlc-hub/grpc/proto"
)

func toStateBlock(blk *types.StateBlock) *pb.StateBlock {
	return &pb.StateBlock{
		Type:           toBlockTypeValue(blk.GetType()),
		Token:          toHashValue(blk.GetToken()),
		Address:        toAddressValue(blk.GetAddress()),
		Balance:        toBalanceValue(blk.GetBalance()),
		Vote:           toBalanceValue(blk.GetVote()),
		Network:        toBalanceValue(blk.GetNetwork()),
		Storage:        toBalanceValue(blk.GetStorage()),
		Oracle:         toBalanceValue(blk.GetOracle()),
		Previous:       toHashValue(blk.GetPrevious()),
		Link:           toHashValue(blk.GetLink()),
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        toHashValue(blk.GetMessage()),
		Data:           blk.GetData(),
		PoVHeight:      blk.PoVHeight,
		Timestamp:      blk.GetTimestamp(),
		Extra:          toHashValue(blk.GetExtra()),
		Representative: toAddressValue(blk.GetRepresentative()),
		PrivateFrom:    blk.PrivateFrom,
		PrivateFor:     blk.PrivateFor,
		PrivateGroupID: blk.PrivateGroupID,
		Work:           toWorkValue(blk.GetWork()),
		Signature:      toSignatureValue(blk.GetSignature()),
	}
}

func toOriginStateBlock(blk *pb.StateBlock) (*types.StateBlock, error) {
	token, err := toOriginHashByValue(blk.GetToken())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(blk.GetAddress())
	if err != nil {
		return nil, err
	}
	pre, err := toOriginHashByValue(blk.GetPrevious())
	if err != nil {
		return nil, err
	}
	link, err := toOriginHashByValue(blk.GetLink())
	if err != nil {
		return nil, err
	}
	message, err := toOriginHashByValue(blk.GetMessage())
	if err != nil {
		return nil, err
	}
	extra, err := toOriginHashByValue(blk.GetExtra())
	if err != nil {
		return nil, err
	}
	rep, err := toOriginAddressByValue(blk.GetRepresentative())
	if err != nil {
		return nil, err
	}
	sign, err := toOriginSignatureByValue(blk.GetSignature())
	if err != nil {
		return nil, err
	}
	return &types.StateBlock{
		Type:           toOriginBlockValue(blk.GetType()),
		Token:          token,
		Address:        addr,
		Balance:        types.Balance{Int: big.NewInt(blk.GetBalance())},
		Vote:           types.ToBalance(types.Balance{Int: big.NewInt(blk.GetVote())}),
		Network:        types.ToBalance(types.Balance{Int: big.NewInt(blk.GetNetwork())}),
		Storage:        types.ToBalance(types.Balance{Int: big.NewInt(blk.GetStorage())}),
		Oracle:         types.ToBalance(types.Balance{Int: big.NewInt(blk.GetOracle())}),
		Previous:       pre,
		Link:           link,
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        types.ToHash(message),
		Data:           blk.GetData(),
		PoVHeight:      blk.GetPoVHeight(),
		Timestamp:      blk.GetTimestamp(),
		Extra:          &extra,
		Representative: rep,
		PrivateFrom:    blk.GetPrivateFrom(),
		PrivateFor:     blk.GetPrivateFor(),
		PrivateGroupID: blk.GetPrivateGroupID(),
		Work:           toOriginWorkByValue(blk.GetWork()),
		Signature:      sign,
	}, nil
}

func toBlockTypeValue(b types.BlockType) string {
	return b.String()
}
func toOriginBlockValue(b string) types.BlockType {
	return types.BlockTypeFromStr(b)
}

func toHashValue(hash types.Hash) string {
	return hash.String()
}
func toOriginHashByValue(hash string) (types.Hash, error) {
	return types.NewHash(hash)
}

func toAddressValue(addr types.Address) string {
	return addr.String()
}
func toOriginAddressByValue(addr string) (types.Address, error) {
	return types.HexToAddress(addr)
}

func toBalanceValue(b types.Balance) int64 {
	if b.Int == nil {
		return types.ZeroBalance.Int64()
	}
	return b.Int64()
}
func toOriginBalanceByValue(b int64) types.Balance {
	return types.Balance{Int: big.NewInt(b)}
}

func toWorkValue(b types.Work) uint64 {
	return uint64(b)
}
func toOriginWorkByValue(v uint64) types.Work {
	return types.Work(v)
}

func toSignatureValue(b types.Signature) string {
	return b.String()
}
func toOriginSignatureByValue(s string) (types.Signature, error) {
	sign, err := types.NewSignature(s)
	if err != nil {
		return types.ZeroSignature, err
	}
	return sign, nil
}
