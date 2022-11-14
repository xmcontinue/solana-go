package sol

import (
	"context"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

var positionV2Len = 32 + 32 + 16 + 4 + 4 + 16 + 8 + 16 + 8 + (24 * 3) + 8

func GetSwapAccountForSwapKeyV2(swapKey solana.PublicKey) (*SwapAccountV2, error) {
	swapAccount := SwapAccountV2{}
	resp, err := GetRpcClient().GetAccountInfo(
		context.Background(),
		swapKey,
	)
	if err != nil {
		return nil, err
	}

	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()[8:]).Decode(&swapAccount)
	if err != nil {
		return nil, err
	}

	return &swapAccount, nil
}

func GetSwapAccountAndPositionsAccountForSwapKeyV2(swapKey solana.PublicKey) (*SwapAccountAndPositionsAccountV2, error) {

	swapAccount, err := GetSwapAccountForSwapKeyV2(swapKey)
	if err != nil {
		return nil, err
	}

	positions, err := GetPositionInfoV2(swapKey)
	if err != nil {
		return nil, err
	}

	swapAccountAndPositionsAccount := SwapAccountAndPositionsAccountV2{
		swapAccount,
		positions,
	}

	return &swapAccountAndPositionsAccount, nil
}
