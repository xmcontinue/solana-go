package event

import (
	"github.com/gagliardetto/solana-go"
)

var CollectEventName = "CollectFeeEvent"

type CollectEvent struct {
	Pool            solana.PublicKey
	Owner           solana.PublicKey
	PositionNftMint solana.PublicKey
	AmountA         uint64
	AmountB         uint64
}
