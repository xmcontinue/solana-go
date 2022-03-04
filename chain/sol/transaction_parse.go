package sol

import (
	"encoding/binary"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
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
	InnerInstructions []solana.CompiledInstruction
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

	for k, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {
		swapRecord, err := t.parseInstructionToSwapCount(
			accountKeys[instruction.ProgramIDIndex].String(),
			instruction.Data,
			instruction.Accounts,
		)
		if err != nil {
			continue
		}

		swapRecord.InnerInstructions, err = getInnerInstructionsForInstructionIndex(k, t.Data.Meta.InnerInstructions)
		if err != nil {
			continue
		}
		if swapRecord.Direction == 1 {
			swapRecord.InnerInstructions[0], swapRecord.InnerInstructions[1] = swapRecord.InnerInstructions[1], swapRecord.InnerInstructions[0]
		}

		t.SwapRecords = append(t.SwapRecords, swapRecord)

	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for k, compiledInstruction := range innerInstruction.Instructions {
			swapRecord, err := t.parseInstructionToSwapCount(
				accountKeys[compiledInstruction.ProgramIDIndex].String(),
				compiledInstruction.Data,
				uint16ListToInt64List(compiledInstruction.Accounts),
			)
			if err != nil {
				continue
			}
			if len(innerInstruction.Instructions) < k+2 {
				continue
			}

			swapRecord.InnerInstructions = []solana.CompiledInstruction{
				innerInstruction.Instructions[k+1],
				innerInstruction.Instructions[k+2],
			}

			if swapRecord.Direction == 1 {
				swapRecord.InnerInstructions[0], swapRecord.InnerInstructions[1] = swapRecord.InnerInstructions[1], swapRecord.InnerInstructions[0]
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
	if swapConfig.TokenA.SwapTokenAccount != accountKeys[instructionAccounts[cremaSwapIndex.TokenAIndex]].String() {
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

	for k, v := range t.SwapRecords {
		t.calculate(k, v.UserCount, v.SwapConfig)

		t.calculate(k, v.TokenCount, v.SwapConfig)

		if v.TokenCount.TokenAVolume.IsZero() {
			continue
		}
		v.Price = v.TokenCount.TokenBVolume.Div(v.TokenCount.TokenAVolume)
	}

	return nil
}

func (t *Tx) calculate(k int, swapCount *SwapCount, config *SwapConfig) {
	var (
		TokenAPostVolume, TokenBPostVolume, TokenAPostBalance, TokenBPostBalance decimal.Decimal
	)

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if swapCount.TokenAIndex == int64(postVal.AccountIndex) {
			TokenAPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}

		if swapCount.TokenBIndex == int64(postVal.AccountIndex) {
			TokenBPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}
	}

	TokenAPostVolume = decimal.NewFromInt(int64(binary.LittleEndian.Uint64(t.SwapRecords[k].InnerInstructions[0].Data[1:9])))
	TokenBPostVolume = decimal.NewFromInt(int64(binary.LittleEndian.Uint64(t.SwapRecords[k].InnerInstructions[1].Data[1:9])))

	swapCount.TokenAVolume = precisionConversion(TokenAPostVolume, int(config.TokenA.Decimal))
	swapCount.TokenBVolume = precisionConversion(TokenBPostVolume, int(config.TokenB.Decimal))
	swapCount.TokenABalance = precisionConversion(TokenAPostBalance, int(config.TokenA.Decimal))
	swapCount.TokenBBalance = precisionConversion(TokenBPostBalance, int(config.TokenB.Decimal))
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

	for k, v := range t.LiquidityRecords {
		t.calculate(k, v.UserCount, v.SwapConfig)

		t.calculate(k, v.TokenCount, v.SwapConfig)
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

func getInnerInstructionsForInstructionIndex(index int, innerInstructions []rpc.InnerInstruction) ([]solana.CompiledInstruction, error) {
	for _, v := range innerInstructions {
		if int(v.Index) == index {
			return v.Instructions[:2], nil
		}
	}
	return nil, errors.RecordNotFound
}
