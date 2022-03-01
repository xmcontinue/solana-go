package sol

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

type Tx struct {
	Data        rpc.GetTransactionResult
	SwapRecords []*SwapRecord
}

// SwapRecord 解析后的有用数据
type SwapRecord struct {
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
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

var (
	cremaSwapIndex          = &SwapIndex{0, 3, 4, 5, 6}
	cremaSwapProgramAddress = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"
)

func NewTx(txData *domain.TxData) *Tx {
	return &Tx{
		Data:        rpc.GetTransactionResult(*txData),
		SwapRecords: make([]*SwapRecord, 0),
	}
}

// ParseTxToSwap 解析TX到swap
func (t *Tx) ParseTxToSwap() error {
	// parser instructions

	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	for _, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {
		// 仅已知的swap address 才可以解析
		if accountKeys[instruction.ProgramIDIndex].String() != cremaSwapProgramAddress {
			continue
		}

		// swap 函数在合约里面的下表是1
		if instruction.Data[0] != 1 {
			continue
		}

		swapConfig, ok := swapConfigMap[accountKeys[instruction.Accounts[cremaSwapIndex.SwapAddressIndex]].String()]
		if !ok {
			return errors.RecordNotFound
		}

		t.SwapRecords = append(t.SwapRecords, &SwapRecord{
			UserOwnerAddress:  accountKeys[0].String(),
			SwapConfig:        swapConfig,
			UserTokenAAddress: accountKeys[instruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String(),
			UserTokenBAddress: accountKeys[instruction.Accounts[cremaSwapIndex.UserTokenBIndex]].String(),
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

			if accountKeys[compiledInstruction.ProgramIDIndex].String() != cremaSwapProgramAddress {
				continue
			}

			// swap 函数在合约里面的下表是1
			if compiledInstruction.Data[0] != 1 {
				continue
			}

			swapConfig, ok := swapConfigMap[accountKeys[compiledInstruction.Accounts[cremaSwapIndex.SwapAddressIndex]].String()]
			if !ok {
				return errors.RecordNotFound
			}

			t.SwapRecords = append(t.SwapRecords, &SwapRecord{
				UserOwnerAddress:  accountKeys[0].String(),
				SwapConfig:        swapConfig,
				UserTokenAAddress: accountKeys[compiledInstruction.Accounts[cremaSwapIndex.UserTokenAIndex]].String(),
				UserTokenBAddress: accountKeys[compiledInstruction.Accounts[cremaSwapIndex.UserTokenBIndex]].String(),
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

	err := t.parseSwapRecord()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// parseSwapRecord 解析出Swap有用的信息
func (t *Tx) parseSwapRecord() error {
	if len(t.SwapRecords) == 0 {
		return errors.RecordNotFound
	}

	for _, v := range t.SwapRecords {
		t.swapCalculate(v, v.UserCount)

		t.swapCalculate(v, v.TokenCount)
	}

	return nil
}

func (t *Tx) swapCalculate(swapRecord *SwapRecord, swapCount *SwapCount) {
	var (
		TokenAPreVolume  decimal.Decimal
		TokenBPreVolume  decimal.Decimal
		TokenAPostVolume decimal.Decimal
		TokenBPostVolume decimal.Decimal
	)

	for _, preVal := range t.Data.Meta.PreTokenBalances {
		if swapCount.TokenAIndex == int64(preVal.AccountIndex) {
			TokenAPreVolume, _ = decimal.NewFromString(preVal.UiTokenAmount.Amount)
			TokenAPreVolume = TokenAPreVolume.Abs()
			continue
		}

		if swapCount.TokenBIndex == int64(preVal.AccountIndex) {
			TokenBPreVolume, _ = decimal.NewFromString(preVal.UiTokenAmount.Amount)
			TokenBPreVolume = TokenBPreVolume.Abs()
			continue
		}
	}

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if swapCount.TokenAIndex == int64(postVal.AccountIndex) {
			TokenAPostVolume, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			TokenAPostVolume = TokenAPostVolume.Abs()
			continue
		}

		if swapCount.TokenBIndex == int64(postVal.AccountIndex) {
			TokenBPostVolume, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			TokenBPostVolume = TokenBPostVolume.Abs()
			continue
		}
	}

	swapCount.TokenAVolume = precisionConversion(TokenAPostVolume.Sub(TokenAPreVolume), int(swapRecord.SwapConfig.TokenA.Decimal))
	swapCount.TokenBVolume = precisionConversion(TokenBPostVolume.Sub(TokenBPreVolume), int(swapRecord.SwapConfig.TokenB.Decimal))
	swapCount.TokenABalance = precisionConversion(TokenAPostVolume, int(swapRecord.SwapConfig.TokenA.Decimal))
	swapCount.TokenBBalance = precisionConversion(TokenBPostVolume, int(swapRecord.SwapConfig.TokenB.Decimal))
}
