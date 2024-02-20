// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/xmcontinue/solana-go"
	ag_format "github.com/xmcontinue/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// RefundPosition is the `refundPosition` instruction.
type RefundPosition struct {

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
	// [5] = [WRITE] mint
	//
	// [6] = [WRITE] mintAtaAccount
	//
	// [7] = [WRITE] userTokenAAccount
	//
	// [8] = [WRITE] userTokenBAccount
	//
	// [9] = [WRITE] swapTokenAAccount
	//
	// [10] = [WRITE] swapTokenBAccount
	//
	// [11] = [] systemProgram
	//
	// [12] = [] tokenProgram
	//
	// [13] = [] associatedTokenProgram
	//
	// [14] = [] rent
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewRefundPositionInstructionBuilder creates a new `RefundPosition` instruction builder.
func NewRefundPositionInstructionBuilder() *RefundPosition {
	nd := &RefundPosition{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 15),
	}
	return nd
}

// SetUserAccount sets the "user" account.
func (inst *RefundPosition) SetUserAccount(user ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(user).WRITE().SIGNER()
	return inst
}

// GetUserAccount gets the "user" account.
func (inst *RefundPosition) GetUserAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSwapAccountAccount sets the "swapAccount" account.
func (inst *RefundPosition) SetSwapAccountAccount(swapAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(swapAccount).WRITE()
	return inst
}

// GetSwapAccountAccount gets the "swapAccount" account.
func (inst *RefundPosition) GetSwapAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetPositionAccountAccount sets the "positionAccount" account.
func (inst *RefundPosition) SetPositionAccountAccount(positionAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(positionAccount).WRITE()
	return inst
}

// GetPositionAccountAccount gets the "positionAccount" account.
func (inst *RefundPosition) GetPositionAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetCremaPositionNftMintAccount sets the "cremaPositionNftMint" account.
func (inst *RefundPosition) SetCremaPositionNftMintAccount(cremaPositionNftMint ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(cremaPositionNftMint).WRITE()
	return inst
}

// GetCremaPositionNftMintAccount gets the "cremaPositionNftMint" account.
func (inst *RefundPosition) GetCremaPositionNftMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetCremaPositionNftTokenAccountAccount sets the "cremaPositionNftTokenAccount" account.
func (inst *RefundPosition) SetCremaPositionNftTokenAccountAccount(cremaPositionNftTokenAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(cremaPositionNftTokenAccount).WRITE()
	return inst
}

// GetCremaPositionNftTokenAccountAccount gets the "cremaPositionNftTokenAccount" account.
func (inst *RefundPosition) GetCremaPositionNftTokenAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetMintAccount sets the "mint" account.
func (inst *RefundPosition) SetMintAccount(mint ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
func (inst *RefundPosition) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

// SetMintAtaAccountAccount sets the "mintAtaAccount" account.
func (inst *RefundPosition) SetMintAtaAccountAccount(mintAtaAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(mintAtaAccount).WRITE()
	return inst
}

// GetMintAtaAccountAccount gets the "mintAtaAccount" account.
func (inst *RefundPosition) GetMintAtaAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}

// SetUserTokenAAccountAccount sets the "userTokenAAccount" account.
func (inst *RefundPosition) SetUserTokenAAccountAccount(userTokenAAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(userTokenAAccount).WRITE()
	return inst
}

// GetUserTokenAAccountAccount gets the "userTokenAAccount" account.
func (inst *RefundPosition) GetUserTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(7)
}

// SetUserTokenBAccountAccount sets the "userTokenBAccount" account.
func (inst *RefundPosition) SetUserTokenBAccountAccount(userTokenBAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(userTokenBAccount).WRITE()
	return inst
}

// GetUserTokenBAccountAccount gets the "userTokenBAccount" account.
func (inst *RefundPosition) GetUserTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(8)
}

// SetSwapTokenAAccountAccount sets the "swapTokenAAccount" account.
func (inst *RefundPosition) SetSwapTokenAAccountAccount(swapTokenAAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[9] = ag_solanago.Meta(swapTokenAAccount).WRITE()
	return inst
}

// GetSwapTokenAAccountAccount gets the "swapTokenAAccount" account.
func (inst *RefundPosition) GetSwapTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(9)
}

// SetSwapTokenBAccountAccount sets the "swapTokenBAccount" account.
func (inst *RefundPosition) SetSwapTokenBAccountAccount(swapTokenBAccount ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[10] = ag_solanago.Meta(swapTokenBAccount).WRITE()
	return inst
}

// GetSwapTokenBAccountAccount gets the "swapTokenBAccount" account.
func (inst *RefundPosition) GetSwapTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(10)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *RefundPosition) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[11] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *RefundPosition) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(11)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *RefundPosition) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[12] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *RefundPosition) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(12)
}

// SetAssociatedTokenProgramAccount sets the "associatedTokenProgram" account.
func (inst *RefundPosition) SetAssociatedTokenProgramAccount(associatedTokenProgram ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[13] = ag_solanago.Meta(associatedTokenProgram)
	return inst
}

// GetAssociatedTokenProgramAccount gets the "associatedTokenProgram" account.
func (inst *RefundPosition) GetAssociatedTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(13)
}

// SetRentAccount sets the "rent" account.
func (inst *RefundPosition) SetRentAccount(rent ag_solanago.PublicKey) *RefundPosition {
	inst.AccountMetaSlice[14] = ag_solanago.Meta(rent)
	return inst
}

// GetRentAccount gets the "rent" account.
func (inst *RefundPosition) GetRentAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(14)
}

func (inst RefundPosition) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_RefundPosition,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst RefundPosition) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RefundPosition) Validate() error {
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
			return errors.New("accounts.Mint is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.MintAtaAccount is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.UserTokenAAccount is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.UserTokenBAccount is not set")
		}
		if inst.AccountMetaSlice[9] == nil {
			return errors.New("accounts.SwapTokenAAccount is not set")
		}
		if inst.AccountMetaSlice[10] == nil {
			return errors.New("accounts.SwapTokenBAccount is not set")
		}
		if inst.AccountMetaSlice[11] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
		if inst.AccountMetaSlice[12] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
		if inst.AccountMetaSlice[13] == nil {
			return errors.New("accounts.AssociatedTokenProgram is not set")
		}
		if inst.AccountMetaSlice[14] == nil {
			return errors.New("accounts.Rent is not set")
		}
	}
	return nil
}

func (inst *RefundPosition) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("RefundPosition")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=15]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("                  user", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("                  swap", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("              position", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("  cremaPositionNftMint", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta(" cremaPositionNftToken", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("                  mint", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(ag_format.Meta("               mintAta", inst.AccountMetaSlice.Get(6)))
						accountsBranch.Child(ag_format.Meta("            userTokenA", inst.AccountMetaSlice.Get(7)))
						accountsBranch.Child(ag_format.Meta("            userTokenB", inst.AccountMetaSlice.Get(8)))
						accountsBranch.Child(ag_format.Meta("            swapTokenA", inst.AccountMetaSlice.Get(9)))
						accountsBranch.Child(ag_format.Meta("            swapTokenB", inst.AccountMetaSlice.Get(10)))
						accountsBranch.Child(ag_format.Meta("         systemProgram", inst.AccountMetaSlice.Get(11)))
						accountsBranch.Child(ag_format.Meta("          tokenProgram", inst.AccountMetaSlice.Get(12)))
						accountsBranch.Child(ag_format.Meta("associatedTokenProgram", inst.AccountMetaSlice.Get(13)))
						accountsBranch.Child(ag_format.Meta("                  rent", inst.AccountMetaSlice.Get(14)))
					})
				})
		})
}

func (obj RefundPosition) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *RefundPosition) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewRefundPositionInstruction declares a new RefundPosition instruction with the provided parameters and accounts.
func NewRefundPositionInstruction(
	// Accounts:
	user ag_solanago.PublicKey,
	swapAccount ag_solanago.PublicKey,
	positionAccount ag_solanago.PublicKey,
	cremaPositionNftMint ag_solanago.PublicKey,
	cremaPositionNftTokenAccount ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	mintAtaAccount ag_solanago.PublicKey,
	userTokenAAccount ag_solanago.PublicKey,
	userTokenBAccount ag_solanago.PublicKey,
	swapTokenAAccount ag_solanago.PublicKey,
	swapTokenBAccount ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
	associatedTokenProgram ag_solanago.PublicKey,
	rent ag_solanago.PublicKey) *RefundPosition {
	return NewRefundPositionInstructionBuilder().
		SetUserAccount(user).
		SetSwapAccountAccount(swapAccount).
		SetPositionAccountAccount(positionAccount).
		SetCremaPositionNftMintAccount(cremaPositionNftMint).
		SetCremaPositionNftTokenAccountAccount(cremaPositionNftTokenAccount).
		SetMintAccount(mint).
		SetMintAtaAccountAccount(mintAtaAccount).
		SetUserTokenAAccountAccount(userTokenAAccount).
		SetUserTokenBAccountAccount(userTokenBAccount).
		SetSwapTokenAAccountAccount(swapTokenAAccount).
		SetSwapTokenBAccountAccount(swapTokenBAccount).
		SetSystemProgramAccount(systemProgram).
		SetTokenProgramAccount(tokenProgram).
		SetAssociatedTokenProgramAccount(associatedTokenProgram).
		SetRentAccount(rent)
}
