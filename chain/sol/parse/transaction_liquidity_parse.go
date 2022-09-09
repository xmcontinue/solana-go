package parse

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

// LiquidityRecord 解析后的liquidity数据
type LiquidityRecord struct {
	UserOwnerAddress      string
	UserTokenAAddress     string
	UserTokenBAddress     string
	ProgramAddress        string
	Direction             int8 // 0为取出,1为质押
	UserCount             *AmountCount
	TokenCount            *AmountCount
	SwapConfig            *domain.SwapConfig
	InnerInstructions     []solana.CompiledInstruction
	InnerInstructionIndex int64
}

func (l *LiquidityRecord) GetSwapConfig() *domain.SwapConfig {
	return l.SwapConfig
}

func (l *LiquidityRecord) GetUserOwnerAccount() string {
	return l.UserOwnerAddress
}

func (l *LiquidityRecord) GetDirection() int8 {
	return l.Direction
}

func (l *LiquidityRecord) GetTokenALiquidityVolume() decimal.Decimal {
	return l.UserCount.TokenAVolume
}

func (l *LiquidityRecord) GetTokenBLiquidityVolume() decimal.Decimal {
	return l.UserCount.TokenBVolume
}

// ParseTxToLiquidity 解析TX到流动性
func (t *Tx) ParseTxToLiquidity() error {
	// parser instructions
	accountKeys := t.TransAction.Message.AccountKeys

	for k, instruction := range t.TransAction.Message.Instructions {

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
				compiledInstruction.Accounts,
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

func (t *Tx) parseInstructionToLiquidityRecord(programAddress string, data []byte, instructionAccounts []uint16) (*LiquidityRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}

	var (
		cremaLiquidityIndex *Index
		direction           int8
	)

	if data[0] == 2 {
		cremaLiquidityIndex = &Index{0, 3, 4, 5, 6}
		direction = 1
	} else if data[0] == 3 {
		cremaLiquidityIndex = &Index{0, 7, 8, 5, 6}
		direction = 0
	} else {
		return nil, errors.New("not liquidity instruction")
	}

	accountKeys := t.TransAction.Message.AccountKeys

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
		UserCount: &AmountCount{
			TokenAIndex: instructionAccounts[cremaLiquidityIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaLiquidityIndex.UserTokenBIndex],
		},
		TokenCount: &AmountCount{
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
