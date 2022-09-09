package parse

import (
	"encoding/binary"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

// SwapRecord 解析后的swap数据
type SwapRecord struct {
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
	ProgramAddress    string
	Direction         int8 // 0为A->B,1为B->A
	Price             decimal.Decimal
	UserCount         *AmountCount
	TokenCount        *AmountCount
	SwapConfig        *domain.SwapConfig
	InnerInstructions []solana.CompiledInstruction
}

type AmountCount struct {
	TokenAIndex   uint16
	TokenBIndex   uint16
	TokenAVolume  decimal.Decimal
	TokenBVolume  decimal.Decimal
	TokenABalance decimal.Decimal
	TokenBBalance decimal.Decimal
}

type Index struct {
	SwapAddressIndex uint16
	UserTokenAIndex  uint16
	UserTokenBIndex  int64
	TokenAIndex      int64
	TokenBIndex      int64
}

// ParseTxToSwap 解析TX到swap
func (t *Tx) ParseTxToSwap() error {
	// parser instructions
	accountKeys := t.TransAction.Message.AccountKeys

	for k, instruction := range t.TransAction.Message.Instructions {
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
				compiledInstruction.Accounts,
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

func (t *Tx) parseInstructionToSwapCount(programAddress string, data []byte, instructionAccounts []uint16) (*SwapRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}

	if data[0] != 1 {
		return nil, errors.New("not swap instruction")
	}

	accountKeys := t.TransAction.Message.AccountKeys

	cremaSwapIndex := &Index{0, 3, 4, 5, 6}

	swapConfig, ok := swapConfigMap[accountKeys[instructionAccounts[cremaSwapIndex.SwapAddressIndex]].String()]
	if !ok {
		return nil, errors.RecordNotFound
	}

	direction := int8(0)
	if swapConfig.TokenA.SwapTokenAccount != accountKeys[instructionAccounts[cremaSwapIndex.TokenAIndex]].String() {
		cremaSwapIndex = &Index{0, 4, 3, 6, 5}
		direction = 1
	}

	return &SwapRecord{
		UserOwnerAddress:  accountKeys[0].String(),
		SwapConfig:        swapConfig,
		UserTokenAAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenAIndex]].String(),
		UserTokenBAddress: accountKeys[instructionAccounts[cremaSwapIndex.UserTokenBIndex]].String(),
		ProgramAddress:    programAddress,
		Direction:         direction,
		UserCount: &AmountCount{
			TokenAIndex: instructionAccounts[cremaSwapIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaSwapIndex.UserTokenBIndex],
		},
		TokenCount: &AmountCount{
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

func (t *Tx) calculate(k int, amountCount *AmountCount, config *domain.SwapConfig) {
	var (
		TokenAPostVolume, TokenBPostVolume, TokenAPostBalance, TokenBPostBalance decimal.Decimal
	)

	for _, postVal := range t.Data.Meta.PostTokenBalances {
		if amountCount.TokenAIndex == postVal.AccountIndex {
			TokenAPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}

		if amountCount.TokenBIndex == postVal.AccountIndex {
			TokenBPostBalance, _ = decimal.NewFromString(postVal.UiTokenAmount.Amount)
			continue
		}
	}

	accounts := t.TransAction.Message.AccountKeys
	var innerInstructions []solana.CompiledInstruction
	if len(t.ClaimRecords) > 0 {
		innerInstructions = t.ClaimRecords[k].InnerInstructions
	} else if len(t.LiquidityRecords) > 0 {
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

	amountCount.TokenAVolume = PrecisionConversion(TokenAPostVolume, int(config.TokenA.Decimal))
	amountCount.TokenBVolume = PrecisionConversion(TokenBPostVolume, int(config.TokenB.Decimal))
	amountCount.TokenABalance = PrecisionConversion(TokenAPostBalance, int(config.TokenA.Decimal))
	amountCount.TokenBBalance = PrecisionConversion(TokenBPostBalance, int(config.TokenB.Decimal))
}

func (sr *SwapRecord) GetVol() decimal.Decimal {
	if sr.Direction == 0 {
		return sr.TokenCount.TokenAVolume
	}
	return sr.TokenCount.TokenBVolume
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

func (sr *SwapRecord) GetSwapConfig() *domain.SwapConfig {
	return sr.SwapConfig
}

func (sr *SwapRecord) GetUserOwnerAccount() string {
	return sr.UserOwnerAddress
}

func (sr *SwapRecord) GetPrice() decimal.Decimal {
	return sr.Price
}

func (sr *SwapRecord) GetTokenAVolume() decimal.Decimal {
	return sr.UserCount.TokenAVolume
}

func (sr *SwapRecord) GetTokenBVolume() decimal.Decimal {
	return sr.UserCount.TokenBVolume
}

func (sr *SwapRecord) GetTokenABalance() decimal.Decimal {
	return sr.TokenCount.TokenABalance
}

func (sr *SwapRecord) GetTokenBBalance() decimal.Decimal {
	return sr.TokenCount.TokenBBalance
}

func (sr *SwapRecord) GetDirection() int8 {
	return sr.Direction
}
