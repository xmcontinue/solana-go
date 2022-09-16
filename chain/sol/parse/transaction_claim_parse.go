package parse

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

var ClaimType = "Claim"

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
	accountKeys := t.TransAction.Message.AccountKeys

	for k, instruction := range t.TransAction.Message.Instructions {
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
				compiledInstruction.Accounts,
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

	err := t.calculateClaim()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (t *Tx) parseInstructionToClaimCount(programAddress string, data []byte, instructionAccounts []uint16) (*ClaimRecord, error) {
	if programAddress != cremaSwapProgramAddress {
		return nil, errors.New("not crema program")
	}

	if data[0] != 4 {
		return nil, errors.New("not claim instruction")
	}

	accountKeys := t.TransAction.Message.AccountKeys

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

// calculateClaim ...
func (t *Tx) calculateClaim() error {
	if len(t.ClaimRecords) == 0 {
		return errors.RecordNotFound
	}

	for k, v := range t.ClaimRecords {
		t.calculate(k, v.UserCount, v.SwapConfig)

		t.calculate(k, v.TokenCount, v.SwapConfig)
	}

	return nil
}

func (c *ClaimRecord) GetSwapConfig() *domain.SwapConfig {
	return c.SwapConfig
}

func (c *ClaimRecord) GetUserOwnerAccount() string {
	return c.UserOwnerAddress
}

func (c *ClaimRecord) GetTokenACollectVolume() decimal.Decimal {
	return c.UserCount.TokenAVolume
}

func (c *ClaimRecord) GetTokenBCollectVolume() decimal.Decimal {
	return c.UserCount.TokenBVolume
}

func (c *ClaimRecord) GetUserAddress() string {
	return c.UserOwnerAddress
}
