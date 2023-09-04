package event

import (
	Uint "github.com/davidminor/uint128"
	ag_solanago "github.com/gagliardetto/solana-go"
)

var IncreaseLiquidityWithFixedTokenEventName = "IncreaseLiquidityWithFixedTokenEvent"

type IncreaseLiquidityWithFixedTokenEvent struct {
	Pool            ag_solanago.PublicKey
	Owner           ag_solanago.PublicKey
	PositionNftMint ag_solanago.PublicKey
	DeltaLiquidity  Uint.Uint128
	AmountA         uint64
	AmountB         uint64
}
