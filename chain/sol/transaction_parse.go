package sol

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

type Tx struct {
	Data             rpc.GetTransactionResult
	SwapRecords      []*SwapRecord
	LiquidityRecords []*LiquidityRecord
}

// SwapRecord 解析后的swap数据
type SwapRecord struct {
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
	ProgramAddress    string
	Direction         int8 // 0为A->B,1为B->A
	Price             decimal.Decimal
	UserCount         *SwapCount
	TokenCount        *SwapCount
	SwapConfig        *SwapConfig
}

type SwapCount struct {
	TokenAIndex   int64
	TokenBIndex   int64
	TokenAVolume  decimal.Decimal
	TokenBVolume  decimal.Decimal
	TokenABalance decimal.Decimal
	TokenBBalance decimal.Decimal
}

type SwapIndex struct {
	SwapAddressIndex int64
	UserTokenAIndex  int64
	UserTokenBIndex  int64
	TokenAIndex      int64
	TokenBIndex      int64
}

// LiquidityRecord 解析后的liquidity数据
type LiquidityRecord struct {
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
	ProgramAddress    string
	Direction         int8 // 0为取出,1为质押
	UserCount         *SwapCount
	TokenCount        *SwapCount
	SwapConfig        *SwapConfig
}

var (
	cremaSwapProgramAddress = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"
)

func NewTx(txData *domain.TxData) *Tx {
	return &Tx{
		Data:             rpc.GetTransactionResult(*txData),
		SwapRecords:      make([]*SwapRecord, 0),
		LiquidityRecords: make([]*LiquidityRecord, 0),
	}
}

// ParseTxToSwap 解析TX到swap
func (t *Tx) ParseTxToSwap() error {
	// parser instructions
	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	for _, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {
		t.parseInstructionToSwapRecord(instruction)
		// 仅已知的swap address 才可以解析
		programAddress := accountKeys[instruction.ProgramIDIndex].String()
		if programAddress != cremaSwapProgramAddress {
			continue
		}

		// swap 函数在合约里面的是01
		if instruction.Data[0] != 1 {
			continue
		}

		cremaSwapIndex := &SwapIndex{0, 3, 4, 5, 6}

		swapConfig, ok := swapConfigMap[accountKeys[instruction.Accounts[cremaSwapIndex.SwapAddressIndex]].String()]
		if !ok {
			return errors.RecordNotFound
		}

		direction := int8(0)
		if swapConfig.TokenA.SwapTokenAccount != accountKeys[instruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String() {
			cremaSwapIndex = &SwapIndex{0, 4, 3, 6, 5}
			direction = 1
		}

		t.SwapRecords = append(t.SwapRecords, &SwapRecord{
			UserOwnerAddress:  accountKeys[0].String(),
			SwapConfig:        swapConfig,
			UserTokenAAddress: accountKeys[instruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String(),
			UserTokenBAddress: accountKeys[instruction.Accounts[cremaSwapIndex.UserTokenBIndex]].String(),
			ProgramAddress:    programAddress,
			Direction:         direction,
			UserCount: &SwapCount{
				TokenAIndex: instruction.Accounts[cremaSwapIndex.UserTokenAIndex],
				TokenBIndex: instruction.Accounts[cremaSwapIndex.UserTokenBIndex],
			},
			TokenCount: &SwapCount{
				TokenAIndex: instruction.Accounts[cremaSwapIndex.TokenAIndex],
				TokenBIndex: instruction.Accounts[cremaSwapIndex.TokenBIndex],
			},
		})
	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for _, compiledInstruction := range innerInstruction.Instructions {

			programAddress := accountKeys[compiledInstruction.ProgramIDIndex].String()
			if programAddress != cremaSwapProgramAddress {
				continue
			}

			// swap 函数在合约里面的是01
			if compiledInstruction.Data[0] != 1 {
				continue
			}
			cremaSwapIndex := &SwapIndex{0, 3, 4, 5, 6}

			swapConfig, ok := swapConfigMap[accountKeys[compiledInstruction.Accounts[cremaSwapIndex.SwapAddressIndex]].String()]
			if !ok {
				return errors.RecordNotFound
			}

			direction := int8(0)
			if swapConfig.TokenA.SwapTokenAccount != accountKeys[compiledInstruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String() {
				cremaSwapIndex = &SwapIndex{0, 4, 3, 6, 5}
				direction = 1
			}

			t.SwapRecords = append(t.SwapRecords, &SwapRecord{
				UserOwnerAddress:  accountKeys[0].String(),
				SwapConfig:        swapConfig,
				UserTokenAAddress: accountKeys[compiledInstruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String(),
				UserTokenBAddress: accountKeys[compiledInstruction.Accounts[cremaSwapIndex.UserTokenBIndex]].String(),
				ProgramAddress:    programAddress,
				Direction:         direction,
				UserCount: &SwapCount{
					TokenAIndex: int64(compiledInstruction.Accounts[cremaSwapIndex.UserTokenAIndex]),
					TokenBIndex: int64(compiledInstruction.Accounts[cremaSwapIndex.UserTokenBIndex]),
				},
				TokenCount: &SwapCount{
					TokenAIndex: int64(compiledInstruction.Accounts[cremaSwapIndex.TokenAIndex]),
					TokenBIndex: int64(compiledInstruction.Accounts[cremaSwapIndex.TokenBIndex]),
				},
			})
		}
	}

	err := t.calculateSwap()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (t *Tx) parseInstructionToSwapRecord(data interface{}) *SwapRecord {
	// solana.CompiledInstruction{}
	// rpc.ParsedInstruction{}
	return &SwapRecord{}
}

// calculateSwaps ...
func (t *Tx) calculateSwap() error {
	if len(t.SwapRecords) == 0 {
		return errors.RecordNotFound
	}

	for _, v := range t.SwapRecords {
		t.calculate(v.UserCount)

		t.calculate(v.TokenCount)

		if v.TokenCount.TokenAVolume.IsZero() {
			continue
		}
		v.Price = precisionConversion(v.TokenCount.TokenBVolume.Div(v.TokenCount.TokenAVolume), int(v.SwapConfig.TokenA.Decimal))
	}

	return nil
}

func (t *Tx) calculate(swapCount *SwapCount) {
	var (
		TokenAPreVolume, TokenBPreVolume, TokenAPostVolume, TokenBPostVolume decimal.Decimal
	)

	for _, preVal := range t.Data.Meta.PreTokenBalances {
		if swapCount.TokenAIndex == int64(preVal.AccountIndex) {
			TokenAPreVolume = decimal.NewFromFloat(*preVal.UiTokenAmount.UiAmount).Abs()
			continue
		}

		if swapCount.TokenBIndex == int64(preVal.AccountIndex) {
			TokenBPreVolume = decimal.NewFromFloat(*preVal.UiTokenAmount.UiAmount).Abs()
			continue
		}
	}

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if swapCount.TokenAIndex == int64(postVal.AccountIndex) {
			TokenAPostVolume = decimal.NewFromFloat(*postVal.UiTokenAmount.UiAmount).Abs()
			continue
		}

		if swapCount.TokenBIndex == int64(postVal.AccountIndex) {
			TokenBPostVolume = decimal.NewFromFloat(*postVal.UiTokenAmount.UiAmount).Abs()
			continue
		}
	}

	swapCount.TokenAVolume = TokenAPostVolume.Sub(TokenAPreVolume)
	swapCount.TokenBVolume = TokenBPostVolume.Sub(TokenBPreVolume)
	swapCount.TokenABalance = TokenAPostVolume
	swapCount.TokenBBalance = TokenBPostVolume
}

func (sr *SwapRecord) GetVol() decimal.Decimal {
	if sr.Direction == 0 {
		return sr.TokenCount.TokenAVolume
	}
	return sr.TokenCount.TokenBVolume
}

// ParseTxToLiquidity 解析TX到流动性
func (t *Tx) ParseTxToLiquidity() error {
	// parser instructions
	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	for _, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {
		// 仅已知的swap address 才可以解析
		programAddress := accountKeys[instruction.ProgramIDIndex].String()
		if programAddress != cremaSwapProgramAddress {
			continue
		}

		// Liquidity 函数在合约里面的是02,03
		if instruction.Data[0] != 2 && instruction.Data[0] != 3 {
			continue
		}

		var (
			cremaLiquidityIndex *SwapIndex
			direction           int8
		)

		if instruction.Data[0] == 02 {
			cremaLiquidityIndex = &SwapIndex{0, 3, 4, 5, 6}
			direction = 1
		} else {
			cremaLiquidityIndex = &SwapIndex{0, 7, 8, 5, 6}
			direction = 0
		}

		swapConfig, ok := swapConfigMap[accountKeys[instruction.Accounts[cremaLiquidityIndex.SwapAddressIndex]].String()]
		if !ok {
			return errors.RecordNotFound
		}

		t.LiquidityRecords = append(t.LiquidityRecords, &LiquidityRecord{
			UserOwnerAddress:  accountKeys[0].String(),
			SwapConfig:        swapConfig,
			UserTokenAAddress: accountKeys[instruction.Accounts[cremaLiquidityIndex.UserTokenAIndex]].String(),
			UserTokenBAddress: accountKeys[instruction.Accounts[cremaLiquidityIndex.UserTokenBIndex]].String(),
			ProgramAddress:    programAddress,
			Direction:         direction,
			UserCount: &SwapCount{
				TokenAIndex: instruction.Accounts[cremaLiquidityIndex.UserTokenAIndex],
				TokenBIndex: instruction.Accounts[cremaLiquidityIndex.UserTokenBIndex],
			},
			TokenCount: &SwapCount{
				TokenAIndex: instruction.Accounts[cremaLiquidityIndex.TokenAIndex],
				TokenBIndex: instruction.Accounts[cremaLiquidityIndex.TokenBIndex],
			},
		})
	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for _, compiledInstruction := range innerInstruction.Instructions {
			programAddress := accountKeys[compiledInstruction.ProgramIDIndex].String()
			if programAddress != cremaSwapProgramAddress {
				continue
			}

			// Liquidity 函数在合约里面的是02,03
			if compiledInstruction.Data[0] != 2 && compiledInstruction.Data[0] != 3 {
				continue
			}

			var (
				cremaLiquidityIndex *SwapIndex
				direction           int8
			)

			if compiledInstruction.Data[0] == 02 {
				cremaLiquidityIndex = &SwapIndex{0, 3, 4, 5, 6}
				direction = 1
			} else {
				cremaLiquidityIndex = &SwapIndex{0, 7, 8, 5, 6}
				direction = 0
			}

			swapConfig, ok := swapConfigMap[accountKeys[compiledInstruction.Accounts[cremaLiquidityIndex.SwapAddressIndex]].String()]
			if !ok {
				return errors.RecordNotFound
			}

			t.LiquidityRecords = append(t.LiquidityRecords, &LiquidityRecord{
				UserOwnerAddress:  accountKeys[0].String(),
				SwapConfig:        swapConfig,
				UserTokenAAddress: accountKeys[compiledInstruction.Accounts[cremaLiquidityIndex.UserTokenAIndex]].String(),
				UserTokenBAddress: accountKeys[compiledInstruction.Accounts[cremaLiquidityIndex.UserTokenBIndex]].String(),
				ProgramAddress:    programAddress,
				Direction:         direction,
				UserCount: &SwapCount{
					TokenAIndex: int64(compiledInstruction.Accounts[cremaLiquidityIndex.UserTokenAIndex]),
					TokenBIndex: int64(compiledInstruction.Accounts[cremaLiquidityIndex.UserTokenBIndex]),
				},
				TokenCount: &SwapCount{
					TokenAIndex: int64(compiledInstruction.Accounts[cremaLiquidityIndex.TokenAIndex]),
					TokenBIndex: int64(compiledInstruction.Accounts[cremaLiquidityIndex.TokenBIndex]),
				},
			})
		}
	}

	err := t.calculateLiquidity()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// calculateLiquidity ...
func (t *Tx) calculateLiquidity() error {
	if len(t.LiquidityRecords) == 0 {
		return errors.RecordNotFound
	}

	for _, v := range t.LiquidityRecords {
		t.calculate(v.UserCount)

		t.calculate(v.TokenCount)
	}

	return nil
}
