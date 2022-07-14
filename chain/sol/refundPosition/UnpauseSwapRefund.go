// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// UnpauseSwapRefund is the `unpauseSwapRefund` instruction.
type UnpauseSwapRefund struct {

	// [0] = [SIGNER] admin
	//
	// [1] = [WRITE] swapAccount
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewUnpauseSwapRefundInstructionBuilder creates a new `UnpauseSwapRefund` instruction builder.
func NewUnpauseSwapRefundInstructionBuilder() *UnpauseSwapRefund {
	nd := &UnpauseSwapRefund{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetAdminAccount sets the "admin" account.
func (inst *UnpauseSwapRefund) SetAdminAccount(admin ag_solanago.PublicKey) *UnpauseSwapRefund {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(admin).SIGNER()
	return inst
}

// GetAdminAccount gets the "admin" account.
func (inst *UnpauseSwapRefund) GetAdminAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSwapAccountAccount sets the "swapAccount" account.
func (inst *UnpauseSwapRefund) SetSwapAccountAccount(swapAccount ag_solanago.PublicKey) *UnpauseSwapRefund {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(swapAccount).WRITE()
	return inst
}

// GetSwapAccountAccount gets the "swapAccount" account.
func (inst *UnpauseSwapRefund) GetSwapAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst UnpauseSwapRefund) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_UnpauseSwapRefund,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst UnpauseSwapRefund) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UnpauseSwapRefund) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Admin is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.SwapAccount is not set")
		}
	}
	return nil
}

func (inst *UnpauseSwapRefund) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UnpauseSwapRefund")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("admin", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta(" swap", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj UnpauseSwapRefund) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *UnpauseSwapRefund) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewUnpauseSwapRefundInstruction declares a new UnpauseSwapRefund instruction with the provided parameters and accounts.
func NewUnpauseSwapRefundInstruction(
	// Accounts:
	admin ag_solanago.PublicKey,
	swapAccount ag_solanago.PublicKey) *UnpauseSwapRefund {
	return NewUnpauseSwapRefundInstructionBuilder().
		SetAdminAccount(admin).
		SetSwapAccountAccount(swapAccount)
}
