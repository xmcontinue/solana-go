package sol

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/palletone/go-palletone/common/uint128"
)

type Position struct {
	NftTokenId           solana.PublicKey
	LowerTick            int32
	UpperTick            int32
	Liquity              decimalU128
	FeeGrowthInsideALast decimalU12816
	FeeGrowthInsideBLast decimalU12816
	TokenAFee            decimalU12816
	TokenBFee            decimalU12816
}

type FarmingPosition struct {
	Wrapper     solana.PublicKey
	Owner       solana.PublicKey
	Bump        uint8
	Index       uint64
	NftVault    solana.PublicKey
	NftMint     solana.PublicKey
	WrapBalance uint64
	Liquity     decimalU128
	LowerTick   int32
	UpperTick   int32
	Hold        bool
	Reserved    uint64
}

type PositionsAccount struct {
	PositionsHead
	Positions    []Position
	PositionsRaw [][]byte
}

type PositionsHead struct {
	SwapVersion  uint8
	TokenSwapKey solana.PublicKey
	AccountType  uint8
	Len          int32
}

type SwapAccount struct {
	Version          uint8
	TokenSwapKey     solana.PublicKey
	AccountType      uint8
	IsInitialized    uint8
	Nonce            uint8
	TokenProgramId   solana.PublicKey
	Manager          solana.PublicKey
	ManagerTokenA    solana.PublicKey
	ManagerTokenB    solana.PublicKey
	SwapTokenA       solana.PublicKey
	SwapTokenB       solana.PublicKey
	TokenAMint       solana.PublicKey
	TokenBMint       solana.PublicKey
	TicksKey         solana.PublicKey
	PositionsKey     solana.PublicKey
	CurveType        uint8
	Fee              decimalU6412
	ManagerFee       decimalU6412
	TickSpace        uint32
	CurrentSqrtPrice decimalU12812
	CurrentLiquity   decimalU128
	FeeGrowthGlobal0 decimalU12816
	FeeGrowthGlobal1 decimalU12816
	ManagerFeeA      decimalU128
	ManagerFeeB      decimalU128
}

type SwapAccountAndPositionsAccount struct {
	SwapAccount      *SwapAccount
	PositionsAccount []*PositionsAccount
}

type decimalU6412 [8]byte
type decimalU128 [16]byte
type decimalU12812 [16]byte
type decimalU12816 [16]byte

const (
	PositionsHeadLen = 38
	PositionLen      = 120
)

func (d *decimalU6412) Val() decimal.Decimal {
	var v uint64
	_ = bin.NewBorshDecoder(d[:]).Decode(&v)
	return decimal.New(int64(v), 0)
}

func (d *decimalU128) Val() decimal.Decimal {
	return byteToUint128(d[:], 0)
}

func (d *decimalU12812) Val() decimal.Decimal {
	return byteToUint128(d[:], -12)
}

func (d *decimalU12816) Val() decimal.Decimal {
	return byteToUint128(d[:], -16)
}

// var v uint64
// _ = bin.NewBorshDecoder(d[:]).Decode(&v)
// return decimal.New(int64(v), -16)

func byteToUint128(b []byte, exp int32) decimal.Decimal {
	return decimal.NewFromBigInt(uint128.FromBytes(b).Big(), exp)
}

func GetSwapAccountForSwapKey(swapKey solana.PublicKey) (*SwapAccount, error) {
	swapAccount := SwapAccount{}
	resp, err := GetRpcClient().GetAccountInfo(
		context.Background(),
		swapKey,
	)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&swapAccount)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return &swapAccount, nil
}

func GetSwapAccountAndPositionsAccountForSwapKey(swapKey solana.PublicKey) (*SwapAccountAndPositionsAccount, error) {

	swapAccount, err := GetSwapAccountForSwapKey(swapKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	positionsAccount, err := GetPositionsAccountForPositionKey(swapAccount.PositionsKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	swapAccountAndPositionsAccount := SwapAccountAndPositionsAccount{
		swapAccount,
		[]*PositionsAccount{
			positionsAccount,
		},
	}

	return &swapAccountAndPositionsAccount, nil
}

func GetSwapAccountAndPositionsAccountForProgramAccounts(swapKey solana.PublicKey) (*SwapAccountAndPositionsAccount, error) {

	swapAccount, err := GetSwapAccountForSwapKey(swapKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	positionsAccounts, err := getPositionsForProgramAccounts(swapKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	swapAccountAndPositionsAccount := SwapAccountAndPositionsAccount{
		swapAccount,
		positionsAccounts,
	}

	return &swapAccountAndPositionsAccount, nil
}

func getPositionsForProgramAccounts(swapKey solana.PublicKey) ([]*PositionsAccount, error) {
	positionsAccounts := make([]*PositionsAccount, 0)
	opt := rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 1,
					Bytes:  swapKey[:],
				},
			},
		},
	}
	programKey, _ := solana.PublicKeyFromBase58("6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319")
	res, err := GetRpcClient().GetProgramAccountsWithOpts(context.Background(), programKey, &opt)
	if err != nil {
		panic(err)
	}
	for _, v := range res {
		if isPositionKeysAccount(v) {
			pa, err := GetPositionsAccountForPositionKey(v.Pubkey)
			if err != nil {
				return nil, err
			}
			positionsAccounts = append(positionsAccounts, pa)
		}
	}
	return positionsAccounts, nil

}

func isPositionKeysAccount(account *rpc.KeyedAccount) bool {
	var v int8
	_ = bin.NewBorshDecoder([]byte{account.Account.Data.GetBinary()[33]}).Decode(&v)
	return v == 2
}

func GetPositionsAccountForPositionKey(positionKey solana.PublicKey) (*PositionsAccount, error) {
	resp, err := GetRpcClient().GetAccountInfo(
		context.Background(),
		positionKey,
	)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	positionsAccount := PositionsAccount{}

	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&positionsAccount.PositionsHead)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	start := int32(PositionsHeadLen)

	for i := int32(0); i < positionsAccount.Len; i++ {
		position, end := Position{}, start+PositionLen
		err = bin.NewBinDecoder(resp.Value.Data.GetBinary()[start:end]).Decode(&position)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		positionsAccount.Positions = append(positionsAccount.Positions, position)
		positionsAccount.PositionsRaw = append(positionsAccount.PositionsRaw, resp.Value.Data.GetBinary()[start:end])

		start += PositionLen
	}

	// for _, v := range positionsAccount.Positions {

	// i := big.Int{}
	// fmt.Println(v.Liquity.Val(), i.SetBytes(v.Liquity[:]).String())
	// fmt.Println(v.NftTokenId.String(), v.Liquity.Val(), v.FeeGrowthInsideALast.Val(), v.FeeGrowthInsideBLast.Val(), v.TokenAFee.Val(), v.TokenBFee.Val())
	// }

	return &positionsAccount, nil
}

func GetPositionWrapperInfo(positionID solana.PublicKey) {
	p := "CPWdCBwzgC2MNQKz7AGAkZH51BskgA1LY9v8RPikQ2x1"
	pKey, _ := solana.PublicKeyFromBase58(p)
	// pm := "8TJqjSU9CqyucngJxUMT2HTEroM5tQdNSGxD881Pjc9G"
	// pmKey, _ := solana.PublicKeyFromBase58(pm)
	k, _, _ := solana.FindProgramAddress([][]byte{[]byte("Position"), positionID.Bytes()}, pKey)
	resp, err := GetRpcClient().GetAccountInfo(
		context.Background(),
		k,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	info := FarmingPosition{}
	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()[8:]).Decode(&info)

	fmt.Println(info.Owner.String(), info.Wrapper.String(), err)
	return
}

func GetUserAddressForTokenKey(tokenKey solana.PublicKey) (string, error) {
	resp, err := GetRpcClient().GetTokenLargestAccounts(
		context.Background(),
		tokenKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		if _, ok := err.(*jsonrpc.RPCError); ok {
			if err.(*jsonrpc.RPCError).Code == -32602 {
				return "", errors.RecordNotFound
			}
		}
		return "", errors.Wrap(err)
	}
	if len(resp.Value) == 0 {
		return "", errors.RecordNotFound
	}
	if resp.Value[0].Amount == "0" {
		return "", errors.RecordNotFound
	}

	info, err := GetRpcClient().GetAccountInfoWithOpts(context.Background(), resp.Value[0].Address, &rpc.GetAccountInfoOpts{
		Commitment: "",
		DataSlice:  nil,
		Encoding:   solana.EncodingJSONParsed,
	})
	if err != nil {
		return "", errors.Wrap(err)
	}
	account := struct {
		Parsed struct {
			Info struct {
				Owner string `json:"owner"`
			} `json:"info"`
		} `json:"parsed"`
	}{}

	err = json.Unmarshal(info.Value.Data.GetRawJSON(), &account)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return account.Parsed.Info.Owner, nil
}

func (sp SwapAccountAndPositionsAccount) CalculateTokenAmount(position *Position) (decimal.Decimal, decimal.Decimal) {
	lowerSqrtPrice := tick2SqrtPrice(position.LowerTick)
	upperSqrtPrice := tick2SqrtPrice(position.UpperTick)

	liquity, currentSqrtPrice := position.Liquity.Val(), sp.SwapAccount.CurrentSqrtPrice.Val()

	if currentSqrtPrice.LessThan(lowerSqrtPrice) {
		amountA := liquity.Div(lowerSqrtPrice).Sub(liquity.Div(upperSqrtPrice))
		return amountA, decimal.Decimal{}
	} else if currentSqrtPrice.GreaterThan(upperSqrtPrice) {
		amountB := liquity.Mul(upperSqrtPrice).Sub(liquity.Mul(lowerSqrtPrice))
		return decimal.Decimal{}, amountB
	} else {
		amountA := liquity.Div(currentSqrtPrice).Sub(liquity.Div(upperSqrtPrice))
		amountB := liquity.Mul(currentSqrtPrice).Sub(liquity.Mul(lowerSqrtPrice))
		return amountA, amountB
	}

	// if currentSqrtPrice.LessThan(lowerSqrtPrice) {
	// 	amountA := parse.FormatFloatCarry(liquity.Div(lowerSqrtPrice).Sub(liquity.Div(upperSqrtPrice)), 0)
	// 	return amountA, decimal.Decimal{}
	// } else if currentSqrtPrice.GreaterThan(upperSqrtPrice) {
	// 	amountB := parse.FormatFloatCarry(liquity.Mul(upperSqrtPrice).Sub(liquity.Mul(lowerSqrtPrice)), 0)
	// 	return decimal.Decimal{}, amountB
	// } else {
	// 	amountA := parse.FormatFloatCarry(liquity.Div(currentSqrtPrice).Sub(liquity.Div(upperSqrtPrice)), 0)
	// 	amountB := parse.FormatFloatCarry(liquity.Mul(currentSqrtPrice).Sub(liquity.Mul(lowerSqrtPrice)), 0)
	// 	return amountA, amountB
	// }
}

func tick2SqrtPrice(tick int32) decimal.Decimal {
	f, _ := decimal.NewFromInt32(tick).Float64()
	return decimal.NewFromFloat(math.Sqrt(math.Pow(1.0001, f)))
}

const (
	PositionV2DataLen = 216
	PositionDataLen   = 216
	// 正式环境
	ProgramIDV2 = "CLMM9tUoggJu2wagPkkqs9eFG4BWhVBZWkP1qv3Sp7tR"
	// 测试环境
	//ProgramIDV2       = "CcLs6shXAUPEi19SGyCeEHU9QhYAWzV2dRpPPNA4aRb7"
	ProgramID = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"
	//swapAccount = "DVPzDyY3Zi6V11R72wJ2r9k2tJXtNRekmSAuAdZ2bFHd"
)

type PositionV2 struct {
	ClmmPool         solana.PublicKey
	PositionNFTMint  solana.PublicKey
	Liquidity        decimalU128
	TickLowerIndex   int32
	TickUpperIndex   int32
	FeeGrowthInsideA decimalU128
	FeeOwedA         uint64
	FeeGrowthInsideB decimalU128
	FeeOwedB         uint64
	RewardInfos      PositionRewards
}

type PositionReward struct {
	GrowthInside decimalU128
	AmountOwed   uint64
}

type PositionRewards [3]PositionReward

func getProgramAccountsOptsV2(swapAccount solana.PublicKey) *rpc.GetProgramAccountsOpts {
	rpcMem := rpc.RPCFilter{
		Memcmp: &rpc.RPCFilterMemcmp{
			Offset: uint64(8),
			Bytes:  swapAccount.Bytes(),
		},
	}

	rpcSize := rpc.RPCFilter{
		DataSize: PositionV2DataLen,
	}
	opt := &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingBase64,
		Filters: []rpc.RPCFilter{
			rpcMem,
			rpcSize,
		},
	}
	return opt
}

func GetPositionInfoV2(swapAccount solana.PublicKey) ([]*PositionV2, error) {
	opt := getProgramAccountsOptsV2(swapAccount)

	result, err := GetRpcClient().GetProgramAccountsWithOpts(context.TODO(), solana.MustPublicKeyFromBase58(ProgramIDV2), opt)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	positionV2s := make([]*PositionV2, 0, len(result))
	for _, v := range result {
		p2 := &PositionV2{}
		err = bin.NewBinDecoder(v.Account.Data.GetBinary()[8:]).Decode(p2)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		positionV2s = append(positionV2s, p2)
	}

	return positionV2s, nil
}

func GetPositionInfo(swapAccount solana.PublicKey) ([]Position, error) {
	opt := getProgramAccountsOpts(swapAccount)

	result, err := GetRpcClient().GetProgramAccountsWithOpts(context.TODO(), solana.MustPublicKeyFromBase58(ProgramID), opt)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	positionsAccounts := make([]*PositionsAccount, 0)
	for _, v := range result {
		if isPositionKeysAccount(v) {
			positionsAccount, err := GetPositionsAccountForPositionKey(v.Pubkey)
			if err != nil {
				return nil, err
			}
			positionsAccounts = append(positionsAccounts, positionsAccount)
		}
	}

	positions := make([]Position, 0, len(result))
	for i := range positionsAccounts {
		positions = append(positions, positionsAccounts[i].Positions...)
	}

	return positions, nil
}

func getProgramAccountsOpts(swapAccount solana.PublicKey) *rpc.GetProgramAccountsOpts {
	opt := &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 1,
					Bytes:  swapAccount.Bytes(),
				},
			},
		},
	}
	return opt
}
