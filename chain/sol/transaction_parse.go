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

const cremaSwapProgramAddress = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"

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

		swapRecord, err := t.parseInstructionToSwapCount(
			accountKeys[instruction.ProgramIDIndex].String(),
			instruction.Data,
			instruction.Accounts,
		)
		if err != nil {
			continue
		}

		t.SwapRecords = append(t.SwapRecords, swapRecord)

	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for _, compiledInstruction := range innerInstruction.Instructions {

			swapRecord, err := t.parseInstructionToSwapCount(
				accountKeys[compiledInstruction.ProgramIDIndex].String(),
				compiledInstruction.Data,
				uint16ListToInt64List(compiledInstruction.Accounts),
			)
			if err != nil {
				continue
			}

			t.SwapRecords = append(t.SwapRecords, swapRecord)

		}
	}

	err := t.calculateSwap()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (t *Tx) parseInstructionToSwapCount(programAddress string, data []byte, instructionAccounts []int64) (*SwapRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}
	if data[0] != 1 {
		return nil, errors.New("not swap instruction")
	}

	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	cremaSwapIndex := &SwapIndex{0, 3, 4, 5, 6}

	swapConfig, ok := swapConfigMap[accountKeys[instructionAccounts[cremaSwapIndex.SwapAddressIndex]].String()]
	if !ok {
		return nil, errors.RecordNotFound
	}

	direction := int8(0)
	if swapConfig.TokenA.SwapTokenAccount != accountKeys[instructionAccounts[cremaSwapIndex.UserTokenAIndex]].String() {
		cremaSwapIndex = &SwapIndex{0, 4, 3, 6, 5}
		direction = 1
	}

	return &SwapRecord{
		UserOwnerAddress:  accountKeys[0].String(),
		SwapConfig:        swapConfig,
		UserTokenAAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenAIndex]].String(),
		UserTokenBAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenBIndex]].String(),
		ProgramAddress:    programAddress,
		Direction:         direction,
		UserCount: &SwapCount{
			TokenAIndex: instructionAccounts[cremaSwapIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaSwapIndex.UserTokenBIndex],
		},
		TokenCount: &SwapCount{
			TokenAIndex: instructionAccounts[cremaSwapIndex.TokenAIndex],
			TokenBIndex: instructionAccounts[cremaSwapIndex.TokenBIndex],
		},
	}, nil
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
		v.Price = v.TokenCount.TokenBVolume.Div(v.TokenCount.TokenAVolume)
	}

	return nil
}

func (t *Tx) calculate(swapCount *SwapCount) {
	var (
		TokenAPreVolume, TokenBPreVolume, TokenAPostVolume, TokenBPostVolume decimal.Decimal
	)

	for _, preVal := range t.Data.Meta.PreTokenBalances {
		if swapCount.TokenAIndex == int64(preVal.AccountIndex) {
			TokenAPreVolume, _ = decimal.NewFromString(preVal.UiTokenAmount.Amount)
			TokenAPreVolume = precisionConversion(TokenAPreVolume.Abs(), int(preVal.UiTokenAmount.Decimals))
			continue
		}

		if swapCount.TokenBIndex == int64(preVal.AccountIndex) {
			TokenBPreVolume, _ = decimal.NewFromString(preVal.UiTokenAmount.Amount)
			TokenBPreVolume = precisionConversion(TokenBPreVolume.Abs(), int(preVal.UiTokenAmount.Decimals))
			continue
		}
	}

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if swapCount.TokenAIndex == int64(postVal.AccountIndex) {
			TokenAPostVolume, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			TokenAPostVolume = precisionConversion(TokenAPostVolume.Abs(), int(postVal.UiTokenAmount.Decimals))
			continue
		}

		if swapCount.TokenBIndex == int64(postVal.AccountIndex) {
			TokenBPostVolume, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			TokenBPostVolume = precisionConversion(TokenBPostVolume.Abs(), int(postVal.UiTokenAmount.Decimals))
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

		liquidityRecord, err := t.parseInstructionToLiquidityRecord(
			accountKeys[instruction.ProgramIDIndex].String(),
			instruction.Data,
			instruction.Accounts,
		)
		if err != nil {
			continue
		}

		t.LiquidityRecords = append(t.LiquidityRecords, liquidityRecord)

	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for _, compiledInstruction := range innerInstruction.Instructions {

			liquidityRecord, err := t.parseInstructionToLiquidityRecord(
				accountKeys[compiledInstruction.ProgramIDIndex].String(),
				compiledInstruction.Data,
				uint16ListToInt64List(compiledInstruction.Accounts),
			)
			if err != nil {
				continue
			}

			t.LiquidityRecords = append(t.LiquidityRecords, liquidityRecord)

		}
	}

	err := t.calculateLiquidity()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (t *Tx) parseInstructionToLiquidityRecord(programAddress string, data []byte, instructionAccounts []int64) (*LiquidityRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}
	if data[0] != 2 && data[0] != 3 {
		return nil, errors.New("not liquidity instruction")
	}

	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	var (
		cremaLiquidityIndex *SwapIndex
		direction           int8
	)

	if data[0] == 02 {
		cremaLiquidityIndex = &SwapIndex{0, 3, 4, 5, 6}
		direction = 1
	} else {
		cremaLiquidityIndex = &SwapIndex{0, 7, 8, 5, 6}
		direction = 0
	}

	swapConfig, ok := swapConfigMap[accountKeys[instructionAccounts[cremaLiquidityIndex.SwapAddressIndex]].String()]
	if !ok {
		return nil, errors.RecordNotFound
	}

	return &LiquidityRecord{
		UserOwnerAddress:  accountKeys[0].String(),
		SwapConfig:        swapConfig,
		UserTokenAAddress: accountKeys[instructionAccounts[cremaLiquidityIndex.UserTokenAIndex]].String(),
		UserTokenBAddress: accountKeys[instructionAccounts[cremaLiquidityIndex.UserTokenBIndex]].String(),
		ProgramAddress:    programAddress,
		Direction:         direction,
		UserCount: &SwapCount{
			TokenAIndex: instructionAccounts[cremaLiquidityIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaLiquidityIndex.UserTokenBIndex],
		},
		TokenCount: &SwapCount{
			TokenAIndex: instructionAccounts[cremaLiquidityIndex.TokenAIndex],
			TokenBIndex: instructionAccounts[cremaLiquidityIndex.TokenBIndex],
		},
	}, nil
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

func uint16ListToInt64List(int16List []uint16) []int64 {
	int64List := make([]int64, 0, len(int16List))
	for _, v := range int16List {
		int64List = append(int64List, int64(v))
	}
	return int64List
}
