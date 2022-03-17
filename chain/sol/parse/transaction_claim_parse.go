package parse

import (
	"git.cplus.link/go/akit/errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

// ClaimRecord 解析后的Claim数据
type ClaimRecord struct {
	UserOwnerAddress  string
	UserTokenAAddress string
	UserTokenBAddress string
	ProgramAddress    string
	UserCount         *AmountCount
	TokenCount        *AmountCount
	SwapConfig        *domain.SwapConfig
	InnerInstructions []solana.CompiledInstruction
}

// ParseTxToClaim 解析TX到Claim
func (t *Tx) ParseTxToClaim() error {
	// parser instructions
	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	for k, instruction := range t.Data.Transaction.GetParsedTransaction().Message.Instructions {
		claimRecord, err := t.parseInstructionToClaimCount(
			accountKeys[instruction.ProgramIDIndex].String(),
			instruction.Data,
			instruction.Accounts,
		)
		if err != nil {
			continue
		}

		claimRecord.InnerInstructions, err = getClaimInnerInstructionsForInstructionIndex(k, t.Data.Meta.InnerInstructions)
		if err != nil {
			continue
		}

		t.ClaimRecords = append(t.ClaimRecords, claimRecord)

	}

	for _, innerInstruction := range t.Data.Meta.InnerInstructions {
		// 仅已知的swap address 才可以解析
		for k, compiledInstruction := range innerInstruction.Instructions {
			claimRecord, err := t.parseInstructionToClaimCount(
				accountKeys[compiledInstruction.ProgramIDIndex].String(),
				compiledInstruction.Data,
				uint16ListToInt64List(compiledInstruction.Accounts),
			)
			if err != nil {
				continue
			}

			if len(innerInstruction.Instructions) < k+1 {
				continue
			}

			claimRecord.InnerInstructions = []solana.CompiledInstruction{
				innerInstruction.Instructions[k+1],
			}

			if len(innerInstruction.Instructions) > k+1 {
				claimRecord.InnerInstructions = append(claimRecord.InnerInstructions, innerInstruction.Instructions[k+2])
			}

			t.ClaimRecords = append(t.ClaimRecords, claimRecord)

		}
	}

	err := t.calculateSwap()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (t *Tx) parseInstructionToClaimCount(programAddress string, data []byte, instructionAccounts []int64) (*ClaimRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}

	if data[0] != 4 {
		return nil, errors.New("not claim instruction")
	}

	accountKeys := t.Data.Transaction.GetParsedTransaction().Message.AccountKeys

	cremaClaimIndex := &Index{0, 7, 8, 5, 6}

	swapConfig, ok := swapConfigMap[accountKeys[instructionAccounts[cremaClaimIndex.SwapAddressIndex]].String()]
	if !ok {
		return nil, errors.RecordNotFound
	}

	return &ClaimRecord{
		UserOwnerAddress:  accountKeys[0].String(),
		SwapConfig:        swapConfig,
		UserTokenAAddress: accountKeys[instructionAccounts[cremaClaimIndex.UserTokenAIndex]].String(),
		UserTokenBAddress: accountKeys[instructionAccounts[cremaClaimIndex.UserTokenBIndex]].String(),
		ProgramAddress:    programAddress,
		UserCount: &AmountCount{
			TokenAIndex: instructionAccounts[cremaClaimIndex.UserTokenAIndex],
			TokenBIndex: instructionAccounts[cremaClaimIndex.UserTokenBIndex],
		},
		TokenCount: &AmountCount{
			TokenAIndex: instructionAccounts[cremaClaimIndex.TokenAIndex],
			TokenBIndex: instructionAccounts[cremaClaimIndex.TokenBIndex],
		},
	}, nil
}

func getClaimInnerInstructionsForInstructionIndex(index int, innerInstructions []rpc.InnerInstruction) ([]solana.CompiledInstruction, error) {
	for _, v := range innerInstructions {
		if int(v.Index) == index {
			return v.Instructions, nil
		}
	}
	return nil, errors.RecordNotFound
}
