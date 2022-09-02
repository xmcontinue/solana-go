package parse

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"

	event "git.cplus.link/crema/backend/chain/event/parser"
	"git.cplus.link/crema/backend/pkg/domain"
)

// SwapRecordV2 解析后的swap数据
type SwapRecordV2 struct {
	SwapAccount       string
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
	ProgramAddress    string
	Direction         int8 // 0为A->B,1为B->A
	AmountIn          decimal.Decimal
	AmountOut         decimal.Decimal
	VaultAAmount      decimal.Decimal
	VaultBAmount      decimal.Decimal
	Price             decimal.Decimal
	SwapConfig        *domain.SwapConfig
}

func (s *SwapRecordV2) GetSwapConfig() *domain.SwapConfig {
	return s.SwapConfig
}

func (s *SwapRecordV2) GetUserOwnerAccount() string {
	return s.UserOwnerAddress
}

func (s *SwapRecordV2) GetPrice() decimal.Decimal {
	return s.Price
}
func (s *SwapRecordV2) GetTokenAVolume() decimal.Decimal {
	return s.AmountIn
}
func (s *SwapRecordV2) GetTokenBVolume() decimal.Decimal {
	return s.AmountOut
}
func (s *SwapRecordV2) GetTokenABalance() decimal.Decimal {
	return s.VaultAAmount
}
func (s *SwapRecordV2) GetTokenBBalance() decimal.Decimal {
	return s.VaultBAmount
}
func (s *SwapRecordV2) GetDirection() int8 {
	return s.Direction
}

func (t *Txv2) createSwapRecord(logMessageEvent event.EventRep) error {
	swap := logMessageEvent.Event.(*event.SwapEvent)

	swapConfig, ok := swapConfigMap[swap.Pool.String()]
	if !ok {
		return nil
	}

	UserTokenA, _, err := solana.FindAssociatedTokenAddress(swap.Owner, solana.MustPublicKeyFromBase58(swapConfig.TokenA.TokenMint))
	if err != nil {
		return errors.Wrap(err)
	}

	UserTokenB, _, err := solana.FindAssociatedTokenAddress(swap.Owner, solana.MustPublicKeyFromBase58(swapConfig.TokenB.TokenMint))
	if err != nil {
		return errors.Wrap(err)
	}

	t.SwapRecords = append(t.SwapRecords, &SwapRecordV2{
		UserOwnerAddress:  swap.Owner.String(),
		UserTokenAAddress: UserTokenA.String(),
		UserTokenBAddress: UserTokenB.String(),
		ProgramAddress:    cremaSwapProgramAddressV2,
		Direction: func() int8 {
			if swap.AToB {
				return 1
			}
			return 0
		}(),
		AmountIn:     PrecisionConversion(decimal.NewFromInt(int64(swap.AmountOut)), int(swapConfig.TokenA.Decimal)), // ToDo  确认顺序
		AmountOut:    PrecisionConversion(decimal.NewFromInt(int64(swap.AmountIn)), int(swapConfig.TokenB.Decimal)),
		VaultAAmount: PrecisionConversion(decimal.NewFromInt(int64(swap.VaultAAmount)), int(swapConfig.TokenA.Decimal)),
		VaultBAmount: PrecisionConversion(decimal.NewFromInt(int64(swap.VaultBAmount)), int(swapConfig.TokenB.Decimal)),
		Price:        decimal.NewFromFloatWithExponent(float64(swap.AmountOut), int32(swapConfig.TokenA.Decimal)).Div(decimal.NewFromFloatWithExponent(float64(swap.AmountIn), int32(swapConfig.TokenA.Decimal))),
		SwapConfig:   swapConfig,
	})

	return nil
}
