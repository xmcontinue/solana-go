package sol

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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
	CurrentLiquity   uint64
	FeeGrowthGlobal0 uint64
	FeeGrowthGlobal1 uint64
	ManagerFeeA      uint64
	ManagerFeeB      uint64
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
	var v uint64
	_ = bin.NewBorshDecoder(d[:]).Decode(&v)
	return decimal.New(int64(v), 0)
}

func (d *decimalU12812) Val() decimal.Decimal {
	var v uint64
	_ = bin.NewBorshDecoder(d[:]).Decode(&v)
	return decimal.New(int64(v), -12)
}

func (d *decimalU12816) Val() decimal.Decimal {
	var v uint64
	_ = bin.NewBorshDecoder(d[:]).Decode(&v)
	return decimal.New(int64(v), -16)
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

func GetPositionsAccountForSwapKey(swapKey solana.PublicKey) (*PositionsAccount, error) {
	swapAccount, err := GetSwapAccountForSwapKey(swapKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return GetPositionsAccountForPositionKey(swapAccount.PositionsKey)
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
	// 	i := big.Int{}
	// 	fmt.Println(v.Liquity.Val(), i.SetBytes(v.Liquity[:]).String())
	// 	fmt.Println(v.NftTokenId.String(), v.Liquity.Val(), v.FeeGrowthInsideALast.Val(), v.FeeGrowthInsideBLast.Val(), v.TokenAFee.Val(), v.TokenBFee.Val())
	// }

	return &positionsAccount, nil
}

func GetUserAddressForTokenKey(tokenKey solana.PublicKey) (string, error) {
	resp, err := GetRpcClient().GetTokenLargestAccounts(
		context.Background(),
		tokenKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", errors.Wrap(err)
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
