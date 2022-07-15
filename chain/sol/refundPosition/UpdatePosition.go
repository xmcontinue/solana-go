// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// UpdatePosition is the `updatePosition` instruction.
type UpdatePosition struct {
	OriginTokenAAmount *uint64 `bin:"optional"`
	OriginTokenBAmount *uint64 `bin:"optional"`
	RefundTokenAAmount *uint64 `bin:"optional"`
	RefundTokenBAmount *uint64 `bin:"optional"`
	CrmAmount          *uint64 `bin:"optional"`

	// [0] = [WRITE, SIGNER] admin
	//
	// [1] = [WRITE] positionAccount
	//
	// [2] = [WRITE] swapAccount
	//
	// [3] = [] cremaPositionNftMint
	//
	// [4] = [] cremaUserPosition
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewUpdatePositionInstructionBuilder creates a new `UpdatePosition` instruction builder.
func NewUpdatePositionInstructionBuilder() *UpdatePosition {
	nd := &UpdatePosition{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetOriginTokenAAmount sets the "originTokenAAmount" parameter.
func (inst *UpdatePosition) SetOriginTokenAAmount(originTokenAAmount uint64) *UpdatePosition {
	inst.OriginTokenAAmount = &originTokenAAmount
	return inst
}

// SetOriginTokenBAmount sets the "originTokenBAmount" parameter.
func (inst *UpdatePosition) SetOriginTokenBAmount(originTokenBAmount uint64) *UpdatePosition {
	inst.OriginTokenBAmount = &originTokenBAmount
	return inst
}

// SetRefundTokenAAmount sets the "refundTokenAAmount" parameter.
func (inst *UpdatePosition) SetRefundTokenAAmount(refundTokenAAmount uint64) *UpdatePosition {
	inst.RefundTokenAAmount = &refundTokenAAmount
	return inst
}

// SetRefundTokenBAmount sets the "refundTokenBAmount" parameter.
func (inst *UpdatePosition) SetRefundTokenBAmount(refundTokenBAmount uint64) *UpdatePosition {
	inst.RefundTokenBAmount = &refundTokenBAmount
	return inst
}

// SetCrmAmount sets the "crmAmount" parameter.
func (inst *UpdatePosition) SetCrmAmount(crmAmount uint64) *UpdatePosition {
	inst.CrmAmount = &crmAmount
	return inst
}

// SetAdminAccount sets the "admin" account.
func (inst *UpdatePosition) SetAdminAccount(admin ag_solanago.PublicKey) *UpdatePosition {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(admin).WRITE().SIGNER()
	return inst
}

// GetAdminAccount gets the "admin" account.
func (inst *UpdatePosition) GetAdminAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetPositionAccountAccount sets the "positionAccount" account.
func (inst *UpdatePosition) SetPositionAccountAccount(positionAccount ag_solanago.PublicKey) *UpdatePosition {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(positionAccount).WRITE()
	return inst
}

// GetPositionAccountAccount gets the "positionAccount" account.
func (inst *UpdatePosition) GetPositionAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetSwapAccountAccount sets the "swapAccount" account.
func (inst *UpdatePosition) SetSwapAccountAccount(swapAccount ag_solanago.PublicKey) *UpdatePosition {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(swapAccount).WRITE()
	return inst
}

// GetSwapAccountAccount gets the "swapAccount" account.
func (inst *UpdatePosition) GetSwapAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetCremaPositionNftMintAccount sets the "cremaPositionNftMint" account.
func (inst *UpdatePosition) SetCremaPositionNftMintAccount(cremaPositionNftMint ag_solanago.PublicKey) *UpdatePosition {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(cremaPositionNftMint)
	return inst
}

// GetCremaPositionNftMintAccount gets the "cremaPositionNftMint" account.
func (inst *UpdatePosition) GetCremaPositionNftMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetCremaUserPositionAccount sets the "cremaUserPosition" account.
func (inst *UpdatePosition) SetCremaUserPositionAccount(cremaUserPosition ag_solanago.PublicKey) *UpdatePosition {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(cremaUserPosition)
	return inst
}

// GetCremaUserPositionAccount gets the "cremaUserPosition" account.
func (inst *UpdatePosition) GetCremaUserPositionAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

func (inst UpdatePosition) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_UpdatePosition,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst UpdatePosition) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *UpdatePosition) Validate() error {
	// Check whether all (required) parameters are set:
	{
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Admin is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.PositionAccount is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.SwapAccount is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.CremaPositionNftMint is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.CremaUserPosition is not set")
		}
	}
	return nil
}

func (inst *UpdatePosition) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("UpdatePosition")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=5]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("OriginTokenAAmount (OPT)", inst.OriginTokenAAmount))
						paramsBranch.Child(ag_format.Param("OriginTokenBAmount (OPT)", inst.OriginTokenBAmount))
						paramsBranch.Child(ag_format.Param("RefundTokenAAmount (OPT)", inst.RefundTokenAAmount))
						paramsBranch.Child(ag_format.Param("RefundTokenBAmount (OPT)", inst.RefundTokenBAmount))
						paramsBranch.Child(ag_format.Param("         CrmAmount (OPT)", inst.CrmAmount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("               admin", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("            position", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("                swap", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("cremaPositionNftMint", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("   cremaUserPosition", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

func (obj UpdatePosition) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `OriginTokenAAmount` param (optional):
	{
		if obj.OriginTokenAAmount == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.OriginTokenAAmount)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `OriginTokenBAmount` param (optional):
	{
		if obj.OriginTokenBAmount == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.OriginTokenBAmount)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `RefundTokenAAmount` param (optional):
	{
		if obj.RefundTokenAAmount == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.RefundTokenAAmount)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `RefundTokenBAmount` param (optional):
	{
		if obj.RefundTokenBAmount == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.RefundTokenBAmount)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `CrmAmount` param (optional):
	{
		if obj.CrmAmount == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.CrmAmount)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (obj *UpdatePosition) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `OriginTokenAAmount` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.OriginTokenAAmount)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `OriginTokenBAmount` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.OriginTokenBAmount)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `RefundTokenAAmount` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.RefundTokenAAmount)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `RefundTokenBAmount` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.RefundTokenBAmount)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `CrmAmount` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.CrmAmount)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewUpdatePositionInstruction declares a new UpdatePosition instruction with the provided parameters and accounts.
func NewUpdatePositionInstruction(
	// Parameters:
	originTokenAAmount uint64,
	originTokenBAmount uint64,
	refundTokenAAmount uint64,
	refundTokenBAmount uint64,
	crmAmount uint64,
	// Accounts:
	admin ag_solanago.PublicKey,
	positionAccount ag_solanago.PublicKey,
	swapAccount ag_solanago.PublicKey,
	cremaPositionNftMint ag_solanago.PublicKey,
	cremaUserPosition ag_solanago.PublicKey) *UpdatePosition {
	return NewUpdatePositionInstructionBuilder().
		SetOriginTokenAAmount(originTokenAAmount).
		SetOriginTokenBAmount(originTokenBAmount).
		SetRefundTokenAAmount(refundTokenAAmount).
		SetRefundTokenBAmount(refundTokenBAmount).
		SetCrmAmount(crmAmount).
		SetAdminAccount(admin).
		SetPositionAccountAccount(positionAccount).
		SetSwapAccountAccount(swapAccount).
		SetCremaPositionNftMintAccount(cremaPositionNftMint).
		SetCremaUserPositionAccount(cremaUserPosition)
}