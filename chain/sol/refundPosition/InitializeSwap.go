// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/xmcontinue/solana-go"
	ag_format "github.com/xmcontinue/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// InitializeSwap is the `initializeSwap` instruction.
type InitializeSwap struct {
	IsCrmReward *bool

	// [0] = [WRITE, SIGNER] admin
	//
	// [1] = [WRITE] swapAccount
	//
	// [2] = [] cremaSwap
	//
	// [3] = [] tokenAMint
	//
	// [4] = [] tokenBMint
	//
	// [5] = [WRITE] tokenAAccount
	//
	// [6] = [WRITE] tokenBAccount
	//
	// [7] = [] tokenProgram
	//
	// [8] = [] associatedTokenProgram
	//
	// [9] = [] systemProgram
	//
	// [10] = [] rent
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewInitializeSwapInstructionBuilder creates a new `InitializeSwap` instruction builder.
func NewInitializeSwapInstructionBuilder() *InitializeSwap {
	nd := &InitializeSwap{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 11),
	}
	return nd
}

// SetIsCrmReward sets the "isCrmReward" parameter.
func (inst *InitializeSwap) SetIsCrmReward(isCrmReward bool) *InitializeSwap {
	inst.IsCrmReward = &isCrmReward
	return inst
}

// SetAdminAccount sets the "admin" account.
func (inst *InitializeSwap) SetAdminAccount(admin ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(admin).WRITE().SIGNER()
	return inst
}

// GetAdminAccount gets the "admin" account.
func (inst *InitializeSwap) GetAdminAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSwapAccountAccount sets the "swapAccount" account.
func (inst *InitializeSwap) SetSwapAccountAccount(swapAccount ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(swapAccount).WRITE()
	return inst
}

// GetSwapAccountAccount gets the "swapAccount" account.
func (inst *InitializeSwap) GetSwapAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetCremaSwapAccount sets the "cremaSwap" account.
func (inst *InitializeSwap) SetCremaSwapAccount(cremaSwap ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(cremaSwap)
	return inst
}

// GetCremaSwapAccount gets the "cremaSwap" account.
func (inst *InitializeSwap) GetCremaSwapAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetTokenAMintAccount sets the "tokenAMint" account.
func (inst *InitializeSwap) SetTokenAMintAccount(tokenAMint ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(tokenAMint)
	return inst
}

// GetTokenAMintAccount gets the "tokenAMint" account.
func (inst *InitializeSwap) GetTokenAMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetTokenBMintAccount sets the "tokenBMint" account.
func (inst *InitializeSwap) SetTokenBMintAccount(tokenBMint ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(tokenBMint)
	return inst
}

// GetTokenBMintAccount gets the "tokenBMint" account.
func (inst *InitializeSwap) GetTokenBMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetTokenAAccountAccount sets the "tokenAAccount" account.
func (inst *InitializeSwap) SetTokenAAccountAccount(tokenAAccount ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(tokenAAccount).WRITE()
	return inst
}

// GetTokenAAccountAccount gets the "tokenAAccount" account.
func (inst *InitializeSwap) GetTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

// SetTokenBAccountAccount sets the "tokenBAccount" account.
func (inst *InitializeSwap) SetTokenBAccountAccount(tokenBAccount ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(tokenBAccount).WRITE()
	return inst
}

// GetTokenBAccountAccount gets the "tokenBAccount" account.
func (inst *InitializeSwap) GetTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *InitializeSwap) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *InitializeSwap) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(7)
}

// SetAssociatedTokenProgramAccount sets the "associatedTokenProgram" account.
func (inst *InitializeSwap) SetAssociatedTokenProgramAccount(associatedTokenProgram ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(associatedTokenProgram)
	return inst
}

// GetAssociatedTokenProgramAccount gets the "associatedTokenProgram" account.
func (inst *InitializeSwap) GetAssociatedTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(8)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *InitializeSwap) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[9] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *InitializeSwap) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(9)
}

// SetRentAccount sets the "rent" account.
func (inst *InitializeSwap) SetRentAccount(rent ag_solanago.PublicKey) *InitializeSwap {
	inst.AccountMetaSlice[10] = ag_solanago.Meta(rent)
	return inst
}

// GetRentAccount gets the "rent" account.
func (inst *InitializeSwap) GetRentAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(10)
}

func (inst InitializeSwap) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_InitializeSwap,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst InitializeSwap) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeSwap) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.IsCrmReward == nil {
			return errors.New("IsCrmReward parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Admin is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.SwapAddress is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.CremaSwap is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.TokenAMint is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.TokenBMint is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.TokenAAccount is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.TokenBAccount is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.AssociatedTokenProgram is not set")
		}
		if inst.AccountMetaSlice[9] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
		if inst.AccountMetaSlice[10] == nil {
			return errors.New("accounts.Rent is not set")
		}
	}
	return nil
}

func (inst *InitializeSwap) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeSwap")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("IsCrmReward", *inst.IsCrmReward))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=11]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("                 admin", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("                  swap", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("             cremaSwap", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("            tokenAMint", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("            tokenBMint", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("                tokenA", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(ag_format.Meta("                tokenB", inst.AccountMetaSlice.Get(6)))
						accountsBranch.Child(ag_format.Meta("          tokenProgram", inst.AccountMetaSlice.Get(7)))
						accountsBranch.Child(ag_format.Meta("associatedTokenProgram", inst.AccountMetaSlice.Get(8)))
						accountsBranch.Child(ag_format.Meta("         systemProgram", inst.AccountMetaSlice.Get(9)))
						accountsBranch.Child(ag_format.Meta("                  rent", inst.AccountMetaSlice.Get(10)))
					})
				})
		})
}

func (obj InitializeSwap) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `IsCrmReward` param:
	err = encoder.Encode(obj.IsCrmReward)
	if err != nil {
		return err
	}
	return nil
}
func (obj *InitializeSwap) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `IsCrmReward`:
	err = decoder.Decode(&obj.IsCrmReward)
	if err != nil {
		return err
	}
	return nil
}

// NewInitializeSwapInstruction declares a new InitializeSwap instruction with the provided parameters and accounts.
func NewInitializeSwapInstruction(
	// Parameters:
	isCrmReward bool,
	// Accounts:
	admin ag_solanago.PublicKey,
	swapAccount ag_solanago.PublicKey,
	cremaSwap ag_solanago.PublicKey,
	tokenAMint ag_solanago.PublicKey,
	tokenBMint ag_solanago.PublicKey,
	tokenAAccount ag_solanago.PublicKey,
	tokenBAccount ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
	associatedTokenProgram ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey,
	rent ag_solanago.PublicKey) *InitializeSwap {
	return NewInitializeSwapInstructionBuilder().
		SetIsCrmReward(isCrmReward).
		SetAdminAccount(admin).
		SetSwapAccountAccount(swapAccount).
		SetCremaSwapAccount(cremaSwap).
		SetTokenAMintAccount(tokenAMint).
		SetTokenBMintAccount(tokenBMint).
		SetTokenAAccountAccount(tokenAAccount).
		SetTokenBAccountAccount(tokenBAccount).
		SetTokenProgramAccount(tokenProgram).
		SetAssociatedTokenProgramAccount(associatedTokenProgram).
		SetSystemProgramAccount(systemProgram).
		SetRentAccount(rent)
}
