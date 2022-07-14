package sol

import (
	"context"

	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/chain/sol/refundPosition"
)

func GetRefundPositions() ([]*refundPosition.Position, error) {
	opt := rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 0,
					Bytes:  []byte{170, 188, 143, 228, 122, 64, 247, 208},
				},
			},
		},
	}
	programKey, _ := solana.PublicKeyFromBase58("5QW9BCx6oZKjSWCVyBZaVU8N4jwtFnged9TsiaXvDj8Q")
	res, err := GetRpcClient().GetProgramAccountsWithOpts(context.Background(), programKey, &opt)
	if err != nil {
		return nil, err
	}

	refundPositions := make([]*refundPosition.Position, 0)
	for _, v := range res {
		info := refundPosition.Position{}

		metadataDecoder := bin.NewDecoderWithEncoding(v.Account.Data.GetBinary(), bin.EncodingBorsh)

		_ = info.UnmarshalWithDecoder(metadataDecoder)

		refundPositions = append(refundPositions, &info)
	}
	return refundPositions, nil
}

func GetRefundSwapPools() (map[string]*refundPosition.Swap, error) {
	opt := rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 0,
					Bytes:  []byte{53, 206, 146, 152, 44, 97, 120, 177},
				},
			},
		},
	}
	programKey, _ := solana.PublicKeyFromBase58("5QW9BCx6oZKjSWCVyBZaVU8N4jwtFnged9TsiaXvDj8Q")
	res, err := GetRpcClient().GetProgramAccountsWithOpts(context.Background(), programKey, &opt)
	if err != nil {
		return nil, err
	}

	swaps := make(map[string]*refundPosition.Swap, 0)
	for _, v := range res {
		info := refundPosition.Swap{}

		metadataDecoder := bin.NewDecoderWithEncoding(v.Account.Data.GetBinary(), bin.EncodingBorsh)

		_ = info.UnmarshalWithDecoder(metadataDecoder)

		swaps[v.Pubkey.String()] = &info
	}
	return swaps, nil
}

type RefundToken struct {
	TokenAAmount decimal.Decimal
	TokenBAmount decimal.Decimal
}

func GetRefundPositionsCount() (map[string]RefundToken, error) {
	positions, err := GetRefundPositions()
	if err != nil {
		return nil, err
	}

	swaps, err := GetRefundSwapPools()
	if err != nil {
		return nil, err
	}

	refundPositionsTvlForSymbol := make(map[string]RefundToken)

	for _, v := range positions {
		swapKey := swaps[v.SwapAccount.String()].CremaSwap
		swapAccount, _ := swapConfigMap[swapKey.String()]
		tokenAAmount, tokenBAmount := parse.PrecisionConversion(decimal.NewFromInt(int64(v.RefundTokenAAmount)), int(swapAccount.TokenA.Decimal)),
			parse.PrecisionConversion(decimal.NewFromInt(int64(v.RefundTokenBAmount)), int(swapAccount.TokenB.Decimal))

		if _, ok := refundPositionsTvlForSymbol[swapKey.String()]; ok {
			refundPositionsTvlForSymbol[swapKey.String()] = RefundToken{
				refundPositionsTvlForSymbol[swapKey.String()].TokenAAmount.Add(tokenAAmount),
				refundPositionsTvlForSymbol[swapKey.String()].TokenBAmount.Add(tokenBAmount),
			}
		} else {
			refundPositionsTvlForSymbol[swapKey.String()] = RefundToken{
				tokenAAmount,
				tokenBAmount,
			}
		}

	}

	return refundPositionsTvlForSymbol, err
}
