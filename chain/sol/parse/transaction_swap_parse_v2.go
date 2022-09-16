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
	GetDirection() int8
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

	if swap.AToB {
		direction = 1
		amountIn = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountIn)), int(swapConfig.TokenA.Decimal))
		amountOut = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountOut)), int(swapConfig.TokenB.Decimal))

		tokenABalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultAAmount)), int(swapConfig.TokenA.Decimal)).Add(amountIn)
		tokenBBalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultBAmount)), int(swapConfig.TokenB.Decimal)).Sub(amountOut)
		//fmt.Println("余额", tokenABalance.String(), tokenBBalance.String())
		//fmt.Println("交易额", amountIn.String(), amountOut.String())
	} else {
		amountIn = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountIn)), int(swapConfig.TokenB.Decimal))
		amountOut = PrecisionConversion(decimal.NewFromInt(int64(swap.AmountOut)), int(swapConfig.TokenA.Decimal))

		tokenABalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultBAmount)), int(swapConfig.TokenB.Decimal)).Add(amountIn)
		tokenBBalance = PrecisionConversion(decimal.NewFromInt(int64(swap.VaultAAmount)), int(swapConfig.TokenA.Decimal)).Sub(amountOut)
		//fmt.Println("余额", tokenABalance.String(), tokenBBalance.String())
		//fmt.Println("交易额", amountIn.String(), amountOut.String())
	}

	t.SwapRecords = append(t.SwapRecords, &SwapRecordV2{
		UserOwnerAddress:  swap.Owner.String(),
		UserTokenAAddress: UserTokenA.String(),
		UserTokenBAddress: UserTokenB.String(),
		ProgramAddress:    cremaSwapProgramAddressV2,
		Direction:         direction,
		AmountIn:          amountIn,
		AmountOut:         amountOut,
		VaultAAmount:      tokenABalance,
		VaultBAmount:      tokenBBalance,
		Price:             PrecisionConversion(decimal.New(int64(swap.AmountOut), 0), int(swapConfig.TokenB.Decimal)).Div(PrecisionConversion(decimal.New(int64(swap.AmountIn), 0), int(swapConfig.TokenA.Decimal))),
		SwapConfig:        swapConfig,
	})

	return nil
}
