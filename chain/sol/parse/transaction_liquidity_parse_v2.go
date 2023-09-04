package parse

import (
	"git.cplus.link/go/akit/util/decimal"

	Uint "github.com/davidminor/uint128"

	event "git.cplus.link/crema/backend/chain/event/parser"
	"git.cplus.link/crema/backend/pkg/domain"
)

type LiquidityRecordV2 struct {
	EventName        string
	SwapConfig       *domain.SwapConfig
	UserOwnerAddress string
	Direction        int8
	DeltaLiquidity   Uint.Uint128
	AmountA          uint64
	AmountB          uint64
	AmountBC         uint64
}

type LiquidityRecordIface interface {
	GetSwapConfig() *domain.SwapConfig
	GetUserOwnerAccount() string
	GetTokenALiquidityVolume() decimal.Decimal
	GetTokenBLiquidityVolume() decimal.Decimal
	GetDirection() int8
	GetUserAddress() string
}

func (l *LiquidityRecordV2) GetSwapConfig() *domain.SwapConfig {
	return l.SwapConfig
}

func (l *LiquidityRecordV2) GetUserOwnerAccount() string {
	return l.UserOwnerAddress
}

func (l *LiquidityRecordV2) GetDirection() int8 {
	return l.Direction
}

func (l *LiquidityRecordV2) GetTokenALiquidityVolume() decimal.Decimal {
	return PrecisionConversion(decimal.NewFromInt(int64(l.AmountA)), int(l.GetSwapConfig().TokenA.Decimal))
}

func (l *LiquidityRecordV2) GetTokenBLiquidityVolume() decimal.Decimal {
	return PrecisionConversion(decimal.NewFromInt(int64(l.AmountB)), int(l.GetSwapConfig().TokenB.Decimal))
}

func (l *LiquidityRecordV2) GetUserAddress() string {
	return l.UserOwnerAddress
}

func (t *Txv2) createIncreaseLiquidityRecord(logMessageEvent event.EventRep) error {
	increaseLiquidityEvent := logMessageEvent.Event.(*event.IncreaseLiquidityEvent)

	swapConfig, ok := swapConfigMap[increaseLiquidityEvent.Pool.String()]
	if !ok {
		return nil
	}

	t.LiquidityRecords = append(t.LiquidityRecords, &LiquidityRecordV2{
		EventName:        event.IncreaseLiquidityEventName,
		SwapConfig:       swapConfig,
		UserOwnerAddress: increaseLiquidityEvent.Owner.String(),
		DeltaLiquidity:   increaseLiquidityEvent.DeltaLiquidity,
		AmountA:          increaseLiquidityEvent.AmountA,
		AmountB:          increaseLiquidityEvent.AmountB,
		Direction:        1,
	})

	return nil
}

func (t *Txv2) createDecreaseLiquidityRecord(logMessageEvent event.EventRep) error {
	decreaseLiquidityEvent := logMessageEvent.Event.(*event.DecreaseLiquidityEvent)

	swapConfig, ok := swapConfigMap[decreaseLiquidityEvent.Pool.String()]
	if !ok {
		return nil
	}

	t.LiquidityRecords = append(t.LiquidityRecords, &LiquidityRecordV2{
		EventName:        event.DecreaseLiquidityEventName,
		SwapConfig:       swapConfig,
		UserOwnerAddress: decreaseLiquidityEvent.Owner.String(),
		DeltaLiquidity:   decreaseLiquidityEvent.DeltaLiquidity,
		AmountA:          decreaseLiquidityEvent.AmountA,
		AmountB:          decreaseLiquidityEvent.AmountB,
		Direction:        0,
	})

	return nil
}
func (t *Txv2) createIncreaseLiquidityWithFixedTokenRecord(logMessageEvent event.EventRep) error {
	fixedTokenEvent := logMessageEvent.Event.(*event.IncreaseLiquidityWithFixedTokenEvent)

	swapConfig, ok := swapConfigMap[fixedTokenEvent.Pool.String()]
	if !ok {
		return nil
	}

	t.LiquidityRecords = append(t.LiquidityRecords, &LiquidityRecordV2{
		EventName:        event.IncreaseLiquidityWithFixedTokenEventName,
		SwapConfig:       swapConfig,
		UserOwnerAddress: fixedTokenEvent.Owner.String(),
		DeltaLiquidity:   fixedTokenEvent.DeltaLiquidity,
		AmountA:          fixedTokenEvent.AmountA,
		AmountB:          fixedTokenEvent.AmountB,
		Direction:        0,
	})

	return nil
}
