package parse

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"

	event "git.cplus.link/crema/backend/chain/event/parser"
	"git.cplus.link/crema/backend/pkg/domain"
)

type SwapRecordIface interface {
	GetSwapConfig() *domain.SwapConfig
	GetUserOwnerAccount() string
	GetPrice() decimal.Decimal
	GetTokenAVolume() decimal.Decimal
	GetTokenBVolume() decimal.Decimal
	GetTokenABalance() decimal.Decimal
	GetTokenBBalance() decimal.Decimal

	GetTokenARefAmount() decimal.Decimal
	GetTokenAFeeAmount() decimal.Decimal
	GetTokenAProtocolAmount() decimal.Decimal

	GetTokenBRefAmount() decimal.Decimal
	GetTokenBFeeAmount() decimal.Decimal
	GetTokenBProtocolAmount() decimal.Decimal

	GetDirection() int8
	GetDirectionByVersion() int8

	GetUserAddress() string
}

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
	RefAmount         decimal.Decimal
	FeeAmount         decimal.Decimal
	ProtocolAmount    decimal.Decimal
	VaultAAmount      decimal.Decimal
	VaultBAmount      decimal.Decimal
	Price             decimal.Decimal
	SwapConfig        *domain.SwapConfig
}

func (s *SwapRecordV2) GetTokenARefAmount() decimal.Decimal {
	return s.RefAmount
}

func (s *SwapRecordV2) GetTokenAFeeAmount() decimal.Decimal {
	return s.FeeAmount
}

func (s *SwapRecordV2) GetTokenAProtocolAmount() decimal.Decimal {
	return s.ProtocolAmount
}

func (s *SwapRecordV2) GetTokenBRefAmount() decimal.Decimal {
	return s.RefAmount
}

func (s *SwapRecordV2) GetTokenBFeeAmount() decimal.Decimal {
	return s.FeeAmount
}

func (s *SwapRecordV2) GetTokenBProtocolAmount() decimal.Decimal {
	return s.ProtocolAmount
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

// GetDirectionByVersion 和v1 版本方向统一
// 因为v1 v2 定义的0 1 方向相反，既
// v1 A->B  为：0，B->A 为：1
// v2 A->B  为：1，B->A 为：0
func (s *SwapRecordV2) GetDirectionByVersion() int8 {
	if s.Direction == 1 {
		return 0
	}
	return 1
}

func (s *SwapRecordV2) GetUserAddress() string {
	return s.UserOwnerAddress
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

	direction, tokenABalance, tokenBBalance, amountIn, amountOut := int8(0), decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero
	refAmount, feeAmount, protocolAmount := decimal.Zero, decimal.Zero, decimal.Zero

	if swap.AToB {
		direction = 1
		amountIn = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountIn)), int(swapConfig.TokenA.Decimal))
		amountOut = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountOut)), int(swapConfig.TokenB.Decimal))

		tokenABalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultAAmount)), int(swapConfig.TokenA.Decimal)).Add(amountIn)
		tokenBBalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultBAmount)), int(swapConfig.TokenB.Decimal)).Sub(amountOut)

		refAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.RefAmount)), int(swapConfig.TokenA.Decimal))
		feeAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.FeeAmount)), int(swapConfig.TokenA.Decimal))
		protocolAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.ProtocolAmount)), int(swapConfig.TokenA.Decimal))
		// fmt.Println("余额", tokenABalance.String(), tokenBBalance.String())
		// fmt.Println("交易额", amountIn.String(), amountOut.String())
	} else {
		amountIn = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountIn)), int(swapConfig.TokenB.Decimal))
		amountOut = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountOut)), int(swapConfig.TokenA.Decimal))

		tokenABalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultBAmount)), int(swapConfig.TokenB.Decimal)).Add(amountIn)
		tokenBBalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultAAmount)), int(swapConfig.TokenA.Decimal)).Sub(amountOut)

		refAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.RefAmount)), int(swapConfig.TokenB.Decimal))
		feeAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.FeeAmount)), int(swapConfig.TokenB.Decimal))
		protocolAmount = PrecisionConversion(decimal.NewFromInt(int64(swap.ProtocolAmount)), int(swapConfig.TokenB.Decimal))
		// fmt.Println("余额", tokenABalance.String(), tokenBBalance.String())
		// fmt.Println("交易额", amountIn.String(), amountOut.String())
	}

	t.SwapRecords = append(t.SwapRecords, &SwapRecordV2{
		UserOwnerAddress:  swap.Owner.String(),
		UserTokenAAddress: UserTokenA.String(),
		UserTokenBAddress: UserTokenB.String(),
		ProgramAddress:    cremaSwapProgramAddressV2,
		Direction:         direction,
		AmountIn:          amountIn,
		AmountOut:         amountOut,
		RefAmount:         refAmount,
		FeeAmount:         feeAmount,
		ProtocolAmount:    protocolAmount,
		VaultAAmount:      tokenABalance,
		VaultBAmount:      tokenBBalance,
		Price:             PrecisionConversion(decimal.New(int64(swap.AmountOut), 0), int(swapConfig.TokenB.Decimal)).Div(PrecisionConversion(decimal.New(int64(swap.AmountIn), 0), int(swapConfig.TokenA.Decimal))),
		SwapConfig:        swapConfig,
	})

	return nil
}
