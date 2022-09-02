// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// RefundPositionV2 is the `refundPositionV2` instruction.
type RefundPositionV2 struct {

	// [0] = [WRITE, SIGNER] user
	//
	// [1] = [WRITE] swapAccount
	//
	// [2] = [WRITE] positionAccount
	//
	// [3] = [WRITE] cremaPositionNftMint
	//
	// [4] = [WRITE] cremaPositionNftTokenAccount
	//
	// [5] = [WRITE] userTokenAAccount
	//
	// [6] = [WRITE] userTokenBAccount
	//
	// [7] = [WRITE] swapTokenAAccount
	//
	// [8] = [WRITE] swapTokenBAccount
	//
	// [9] = [] tokenProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewRefundPositionV2InstructionBuilder creates a new `RefundPositionV2` instruction builder.
func NewRefundPositionV2InstructionBuilder() *RefundPositionV2 {
	nd := &RefundPositionV2{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 10),
	}
	return nd
}

// SetUserAccount sets the "user" account.
func (inst *RefundPositionV2) SetUserAccount(user ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(user).WRITE().SIGNER()
	return inst
}

// GetUserAccount gets the "user" account.
func (inst *RefundPositionV2) GetUserAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSwapAccountAccount sets the "swapAccount" account.
func (inst *RefundPositionV2) SetSwapAccountAccount(swapAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(swapAccount).WRITE()
	return inst
}

// GetSwapAccountAccount gets the "swapAccount" account.
func (inst *RefundPositionV2) GetSwapAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetPositionAccountAccount sets the "positionAccount" account.
func (inst *RefundPositionV2) SetPositionAccountAccount(positionAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(positionAccount).WRITE()
	return inst
}

// GetPositionAccountAccount gets the "positionAccount" account.
func (inst *RefundPositionV2) GetPositionAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetCremaPositionNftMintAccount sets the "cremaPositionNftMint" account.
func (inst *RefundPositionV2) SetCremaPositionNftMintAccount(cremaPositionNftMint ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(cremaPositionNftMint).WRITE()
	return inst
}

// GetCremaPositionNftMintAccount gets the "cremaPositionNftMint" account.
func (inst *RefundPositionV2) GetCremaPositionNftMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetCremaPositionNftTokenAccountAccount sets the "cremaPositionNftTokenAccount" account.
func (inst *RefundPositionV2) SetCremaPositionNftTokenAccountAccount(cremaPositionNftTokenAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(cremaPositionNftTokenAccount).WRITE()
	return inst
}

// GetCremaPositionNftTokenAccountAccount gets the "cremaPositionNftTokenAccount" account.
func (inst *RefundPositionV2) GetCremaPositionNftTokenAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetUserTokenAAccountAccount sets the "userTokenAAccount" account.
func (inst *RefundPositionV2) SetUserTokenAAccountAccount(userTokenAAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(userTokenAAccount).WRITE()
	return inst
}

// GetUserTokenAAccountAccount gets the "userTokenAAccount" account.
func (inst *RefundPositionV2) GetUserTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

// SetUserTokenBAccountAccount sets the "userTokenBAccount" account.
func (inst *RefundPositionV2) SetUserTokenBAccountAccount(userTokenBAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(userTokenBAccount).WRITE()
	return inst
}

// GetUserTokenBAccountAccount gets the "userTokenBAccount" account.
func (inst *RefundPositionV2) GetUserTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}

// SetSwapTokenAAccountAccount sets the "swapTokenAAccount" account.
func (inst *RefundPositionV2) SetSwapTokenAAccountAccount(swapTokenAAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(swapTokenAAccount).WRITE()
	return inst
}

// GetSwapTokenAAccountAccount gets the "swapTokenAAccount" account.
func (inst *RefundPositionV2) GetSwapTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(7)
}

// SetSwapTokenBAccountAccount sets the "swapTokenBAccount" account.
func (inst *RefundPositionV2) SetSwapTokenBAccountAccount(swapTokenBAccount ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(swapTokenBAccount).WRITE()
	return inst
}

// GetSwapTokenBAccountAccount gets the "swapTokenBAccount" account.
func (inst *RefundPositionV2) GetSwapTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(8)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *RefundPositionV2) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *RefundPositionV2 {
	inst.AccountMetaSlice[9] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *RefundPositionV2) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(9)
}

func (inst RefundPositionV2) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_RefundPositionV2,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst RefundPositionV2) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RefundPositionV2) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.User is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.SwapAddress is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.PositionAccount is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.CremaPositionNftMint is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.CremaPositionNftTokenAccount is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.UserTokenAAccount is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.UserTokenBAccount is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.SwapTokenAAccount is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.SwapTokenBAccount is not set")
		}
		if inst.AccountMetaSlice[9] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
	}
	return nil
}

func (inst *RefundPositionV2) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("RefundPositionV2")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=10]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("                 user", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("                 swap", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("             position", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta(" cremaPositionNftMint", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("cremaPositionNftToken", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("           userTokenA", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(ag_format.Meta("           userTokenB", inst.AccountMetaSlice.Get(6)))
						accountsBranch.Child(ag_format.Meta("           swapTokenA", inst.AccountMetaSlice.Get(7)))
						accountsBranch.Child(ag_format.Meta("           swapTokenB", inst.AccountMetaSlice.Get(8)))
						accountsBranch.Child(ag_format.Meta("         tokenProgram", inst.AccountMetaSlice.Get(9)))
					})
				})
		})
}

func (obj RefundPositionV2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *RefundPositionV2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewRefundPositionV2Instruction declares a new RefundPositionV2 instruction with the provided parameters and accounts.
func NewRefundPositionV2Instruction(
	// Accounts:
	user ag_solanago.PublicKey,
	swapAccount ag_solanago.PublicKey,
	positionAccount ag_solanago.PublicKey,
	cremaPositionNftMint ag_solanago.PublicKey,
	cremaPositionNftTokenAccount ag_solanago.PublicKey,
	userTokenAAccount ag_solanago.PublicKey,
	userTokenBAccount ag_solanago.PublicKey,
	swapTokenAAccount ag_solanago.PublicKey,
	swapTokenBAccount ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey) *RefundPositionV2 {
	return NewRefundPositionV2InstructionBuilder().
		SetUserAccount(user).
		SetSwapAccountAccount(swapAccount).
		SetPositionAccountAccount(positionAccount).
		SetCremaPositionNftMintAccount(cremaPositionNftMint).
		SetCremaPositionNftTokenAccountAccount(cremaPositionNftTokenAccount).
		SetUserTokenAAccountAccount(userTokenAAccount).
		SetUserTokenBAccountAccount(userTokenBAccount).
		SetSwapTokenAAccountAccount(swapTokenAAccount).
		SetSwapTokenBAccountAccount(swapTokenBAccount).
		SetTokenProgramAccount(tokenProgram)
}
