package event

import (
	ag_solanago "github.com/xmcontinue/solana-go"
)

var SwapEventName = "SwapEvent"

type SwapEvent struct {
	Pool           ag_solanago.PublicKey
	Owner          ag_solanago.PublicKey
	Partner        ag_solanago.PublicKey
	AToB           bool
	AmountIn       uint64
	AmountOut      uint64
	RefAmount      uint64
	FeeAmount      uint64
	ProtocolAmount uint64
	VaultAAmount   uint64
	VaultBAmount   uint64
}
