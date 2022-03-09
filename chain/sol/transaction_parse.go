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
	UserCount         *TxCount
	TokenCount        *TxCount
	SwapConfig        *domain.SwapConfig
	InnerInstructions []solana.CompiledInstruction
}

type TxCount struct {
	TokenAIndex   int64
	TokenBIndex   int64
	TokenAVolume  decimal.Decimal
	TokenBVolume  decimal.Decimal
	TokenABalance decimal.Decimal
	TokenBBalance decimal.Decimal
}

type TxIndex struct {
	SwapAddressIndex int64
	UserTokenAIndex  int64
	UserTokenBIndex  int64
	TokenAIndex      int64
	TokenBIndex      int64
}

// LiquidityRecord 解析后的liquidity数据
type LiquidityRecord struct {
	UserOwnerAddress      string
	UserTokenAAddress     string
	UserTokenBAddress     string
	ProgramAddress        string
	Direction             int8 // 0为取出,1为质押
	UserCount             *TxCount
	TokenCount            *TxCount
	SwapConfig            *domain.SwapConfig
	InnerInstructions     []solana.CompiledInstruction
	InnerInstructionIndex int64
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

		swapRecord.InnerInstructions, err = getSwapInnerInstructionsForInstructionIndex(k, t.Data.Meta.InnerInstructions)
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

	cremaSwapIndex := &TxIndex{0, 3, 4, 5, 6}

	swapConfig, ok := swapConfigMap[accountKeys[instructionAccounts[cremaSwapIndex.SwapAddressIndex]].String()]
	if !ok {
		return nil, errors.RecordNotFound
	}

	direction := int8(0)
	if swapConfig.TokenA.SwapTokenAccount != accountKeys[instructionAccounts[cremaSwapIndex.TokenAIndex]].String() {
		cremaSwapIndex = &TxIndex{0, 4, 3, 6, 5}
		direction = 1
	}

	return &SwapRecord{
		UserOwnerAddress:  accountKeys[0].String(),
		SwapConfig:        swapConfig,
		UserTokenAAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenAIndex]].String(),
		UserTokenBAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenBIndex]].String(),
		ProgramAddress:    programAddress,
		Direction:         direction,
		UserCount: &TxCount{
			TokenAIndex: instructionAccounts[cremaSwapIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaSwapIndex.UserTokenBIndex],
		},
		TokenCount: &TxCount{
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

func (t *Tx) calculate(k int, txCount *TxCount, config *domain.SwapConfig) {
	var (
		TokenAPostVolume, TokenBPostVolume, TokenAPostBalance, TokenBPostBalance decimal.Decimal
	)

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if txCount.TokenAIndex == int64(postVal.AccountIndex) {
			TokenAPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}

		if txCount.TokenBIndex == int64(postVal.AccountIndex) {
			TokenBPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}
	}

	accounts := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys
	var innerInstructions []solana.CompiledInstruction
	if len(t.LiquidityRecords) > 0 {
		innerInstructions = t.LiquidityRecords[k].InnerInstructions
	} else {
		innerInstructions = t.SwapRecords[k].InnerInstructions
	}

	for _, v := range innerInstructions {
		if accounts[v.Accounts[0]].String() == config.TokenA.SwapTokenAccount ||
			accounts[v.Accounts[1]].String() == config.TokenA.SwapTokenAccount {
			TokenAPostVolume = TokenAPostVolume.Add(decimal.NewFromInt(int64(binary.LittleEndian.Uint64(v.Data[1:9]))))
		} else {
			TokenBPostVolume = TokenBPostVolume.Add(decimal.NewFromInt(int64(binary.LittleEndian.Uint64(v.Data[1:9]))))
		}
	}

	txCount.TokenAVolume = PrecisionConversion(TokenAPostVolume, int(config.TokenA.Decimal))
	txCount.TokenBVolume = PrecisionConversion(TokenBPostVolume, int(config.TokenB.Decimal))
	txCount.TokenABalance = PrecisionConversion(TokenAPostBalance, int(config.TokenA.Decimal))
	txCount.TokenBBalance = PrecisionConversion(TokenBPostBalance, int(config.TokenB.Decimal))
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

	for k, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {

		liquidityRecord, err := t.parseInstructionToLiquidityRecord(
			accountKeys[instruction.ProgramIDIndex].String(),
			instruction.Data,
			instruction.Accounts,
		)
		if err != nil {
			continue
		}

		if liquidityRecord.Direction == 1 {
			liquidityRecord.InnerInstructions, err = getDepositInnerInstructionsForInstructionIndex(k, t.Data.Meta.InnerInstructions)
		} else {
			liquidityRecord.InnerInstructions, err = getWithdrawInnerInstructionsForInstructionIndex(k, t.Data.Meta.InnerInstructions)
		}
		if err != nil {
			continue
		}

		t.LiquidityRecords = append(t.LiquidityRecords, liquidityRecord)

	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for k, compiledInstruction := range innerInstruction.Instructions {

			liquidityRecord, err := t.parseInstructionToLiquidityRecord(
				accountKeys[compiledInstruction.ProgramIDIndex].String(),
				compiledInstruction.Data,
				uint16ListToInt64List(compiledInstruction.Accounts),
			)
			if err != nil {
				continue
			}

			if liquidityRecord.Direction == 1 {
				jumpNum := 0
				if innerInstruction.Instructions[k+1].Data[0] != 3 {
					if len(innerInstruction.Instructions) < k+jumpNum+1 {
						continue
					}
					jumpNum = 2
				}

				if innerInstruction.Instructions[k+jumpNum+1].Data[0] == 3 {
					liquidityRecord.InnerInstructions[0] = innerInstruction.Instructions[k+jumpNum+1]
				}

				if len(innerInstruction.Instructions) >= k+jumpNum+2 {
					if innerInstruction.Instructions[k+jumpNum+2].Data[0] == 3 {
						liquidityRecord.InnerInstructions[1] = innerInstruction.Instructions[k+jumpNum+2]
					}
				}
			} else {
				for _, v := range innerInstruction.Instructions[k:] {
					if v.Data[0] == 3 {
						liquidityRecord.InnerInstructions = append(liquidityRecord.InnerInstructions, v)
					}
				}
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

	var (
		cremaLiquidityIndex *TxIndex
		direction           int8
	)

	if data[0] == 2 {
		cremaLiquidityIndex = &TxIndex{0, 3, 4, 5, 6}
		direction = 1
	} else if data[0] == 3 {
		cremaLiquidityIndex = &TxIndex{0, 7, 8, 5, 6}
		direction = 0
	} else {
		return nil, errors.New("not liquidity instruction")
	}

	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

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
		UserCount: &TxCount{
			TokenAIndex: instructionAccounts[cremaLiquidityIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaLiquidityIndex.UserTokenBIndex],
		},
		TokenCount: &TxCount{
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

func getSwapInnerInstructionsForInstructionIndex(index int, innerInstructions []rpc.InnerInstruction) ([]solana.CompiledInstruction, error) {
	for _, v := range innerInstructions {
		if int(v.Index) == index {
			return v.Instructions[:2], nil
		}
	}
	return nil, errors.RecordNotFound
}

func getDepositInnerInstructionsForInstructionIndex(index int, innerInstructions []rpc.InnerInstruction) ([]solana.CompiledInstruction, error) {
	for _, v := range innerInstructions {
		if int(v.Index) == index {
			l := len(v.Instructions)
			if l <= 2 {
				return v.Instructions[:l], nil
			}
			if l <= 4 {
				return v.Instructions[2:l], nil
			}
		}
	}
	return nil, errors.RecordNotFound
}

func getWithdrawInnerInstructionsForInstructionIndex(index int, innerInstructions []rpc.InnerInstruction) ([]solana.CompiledInstruction, error) {
	for _, v := range innerInstructions {
		if int(v.Index) == index {
			return v.Instructions[1:], nil
		}
	}
	return nil, errors.RecordNotFound
}
