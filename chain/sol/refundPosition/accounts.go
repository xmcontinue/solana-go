// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import (
	"fmt"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type Position struct {
	SwapAccount          ag_solanago.PublicKey
	PositionMint         ag_solanago.PublicKey
	Bump                 uint8
	NftTokenId           ag_solanago.PublicKey
	LowerTick            int32
	UpperTick            int32
	Liquity              ag_binary.Uint128
	FeeGrowthInsideALast ag_binary.Uint128
	FeeGrowthInsideBLast ag_binary.Uint128
	TokenAFee            ag_binary.Uint128
	TokenBFee            ag_binary.Uint128
	OriginTokenAAmount   uint64
	OriginTokenBAmount   uint64
	RefundTokenAAmount   uint64
	RefundTokenBAmount   uint64
	CrmAmount            uint64
	IsRefunded           bool
}

var PositionDiscriminator = [8]byte{170, 188, 143, 228, 122, 64, 247, 208}

func (obj Position) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(PositionDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `SwapAddress` param:
	err = encoder.Encode(obj.SwapAccount)
	if err != nil {
		return err
	}
	// Serialize `PositionMint` param:
	err = encoder.Encode(obj.PositionMint)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `NftTokenId` param:
	err = encoder.Encode(obj.NftTokenId)
	if err != nil {
		return err
	}
	// Serialize `LowerTick` param:
	err = encoder.Encode(obj.LowerTick)
	if err != nil {
		return err
	}
	// Serialize `UpperTick` param:
	err = encoder.Encode(obj.UpperTick)
	if err != nil {
		return err
	}
	// Serialize `Liquity` param:
	err = encoder.Encode(obj.Liquity)
	if err != nil {
		return err
	}
	// Serialize `FeeGrowthInsideALast` param:
	err = encoder.Encode(obj.FeeGrowthInsideALast)
	if err != nil {
		return err
	}
	// Serialize `FeeGrowthInsideBLast` param:
	err = encoder.Encode(obj.FeeGrowthInsideBLast)
	if err != nil {
		return err
	}
	// Serialize `TokenAFee` param:
	err = encoder.Encode(obj.TokenAFee)
	if err != nil {
		return err
	}
	// Serialize `TokenBFee` param:
	err = encoder.Encode(obj.TokenBFee)
	if err != nil {
		return err
	}
	// Serialize `OriginTokenAAmount` param:
	err = encoder.Encode(obj.OriginTokenAAmount)
	if err != nil {
		return err
	}
	// Serialize `OriginTokenBAmount` param:
	err = encoder.Encode(obj.OriginTokenBAmount)
	if err != nil {
		return err
	}
	// Serialize `RefundTokenAAmount` param:
	err = encoder.Encode(obj.RefundTokenAAmount)
	if err != nil {
		return err
	}
	// Serialize `RefundTokenBAmount` param:
	err = encoder.Encode(obj.RefundTokenBAmount)
	if err != nil {
		return err
	}
	// Serialize `CrmAmount` param:
	err = encoder.Encode(obj.CrmAmount)
	if err != nil {
		return err
	}
	// Serialize `IsRefunded` param:
	err = encoder.Encode(obj.IsRefunded)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Position) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(PositionDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[170 188 143 228 122 64 247 208]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `SwapAddress`:
	err = decoder.Decode(&obj.SwapAccount)
	if err != nil {
		return err
	}
	// Deserialize `PositionMint`:
	err = decoder.Decode(&obj.PositionMint)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `NftTokenId`:
	err = decoder.Decode(&obj.NftTokenId)
	if err != nil {
		return err
	}
	// Deserialize `LowerTick`:
	err = decoder.Decode(&obj.LowerTick)
	if err != nil {
		return err
	}
	// Deserialize `UpperTick`:
	err = decoder.Decode(&obj.UpperTick)
	if err != nil {
		return err
	}
	// Deserialize `Liquity`:
	err = decoder.Decode(&obj.Liquity)
	if err != nil {
		return err
	}
	// Deserialize `FeeGrowthInsideALast`:
	err = decoder.Decode(&obj.FeeGrowthInsideALast)
	if err != nil {
		return err
	}
	// Deserialize `FeeGrowthInsideBLast`:
	err = decoder.Decode(&obj.FeeGrowthInsideBLast)
	if err != nil {
		return err
	}
	// Deserialize `TokenAFee`:
	err = decoder.Decode(&obj.TokenAFee)
	if err != nil {
		return err
	}
	// Deserialize `TokenBFee`:
	err = decoder.Decode(&obj.TokenBFee)
	if err != nil {
		return err
	}
	// Deserialize `OriginTokenAAmount`:
	err = decoder.Decode(&obj.OriginTokenAAmount)
	if err != nil {
		return err
	}
	// Deserialize `OriginTokenBAmount`:
	err = decoder.Decode(&obj.OriginTokenBAmount)
	if err != nil {
		return err
	}
	// Deserialize `RefundTokenAAmount`:
	err = decoder.Decode(&obj.RefundTokenAAmount)
	if err != nil {
		return err
	}
	// Deserialize `RefundTokenBAmount`:
	err = decoder.Decode(&obj.RefundTokenBAmount)
	if err != nil {
		return err
	}
	// Deserialize `CrmAmount`:
	err = decoder.Decode(&obj.CrmAmount)
	if err != nil {
		return err
	}
	// Deserialize `IsRefunded`:
	err = decoder.Decode(&obj.IsRefunded)
	if err != nil {
		return err
	}
	return nil
}

type Swap struct {
	CremaSwap        ag_solanago.PublicKey
	Admin            ag_solanago.PublicKey
	Bump             [1]uint8
	TokenAMint       ag_solanago.PublicKey
	TokenBMint       ag_solanago.PublicKey
	TokenAAccount    ag_solanago.PublicKey
	TokenBAccount    ag_solanago.PublicKey
	CurrentSqrtPrice ag_binary.Uint128
	IsPause          bool
	PositionNum      uint64
	IsCrmReward      bool
}

var SwapDiscriminator = [8]byte{53, 206, 146, 152, 44, 97, 120, 177}

func (obj Swap) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(SwapDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `CremaSwap` param:
	err = encoder.Encode(obj.CremaSwap)
	if err != nil {
		return err
	}
	// Serialize `Admin` param:
	err = encoder.Encode(obj.Admin)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `TokenAMint` param:
	err = encoder.Encode(obj.TokenAMint)
	if err != nil {
		return err
	}
	// Serialize `TokenBMint` param:
	err = encoder.Encode(obj.TokenBMint)
	if err != nil {
		return err
	}
	// Serialize `TokenAAccount` param:
	err = encoder.Encode(obj.TokenAAccount)
	if err != nil {
		return err
	}
	// Serialize `TokenBAccount` param:
	err = encoder.Encode(obj.TokenBAccount)
	if err != nil {
		return err
	}
	// Serialize `CurrentSqrtPrice` param:
	err = encoder.Encode(obj.CurrentSqrtPrice)
	if err != nil {
		return err
	}
	// Serialize `IsPause` param:
	err = encoder.Encode(obj.IsPause)
	if err != nil {
		return err
	}
	// Serialize `PositionNum` param:
	err = encoder.Encode(obj.PositionNum)
	if err != nil {
		return err
	}
	// Serialize `IsCrmReward` param:
	err = encoder.Encode(obj.IsCrmReward)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Swap) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(SwapDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[53 206 146 152 44 97 120 177]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `CremaSwap`:
	err = decoder.Decode(&obj.CremaSwap)
	if err != nil {
		return err
	}
	// Deserialize `Admin`:
	err = decoder.Decode(&obj.Admin)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `TokenAMint`:
	err = decoder.Decode(&obj.TokenAMint)
	if err != nil {
		return err
	}
	// Deserialize `TokenBMint`:
	err = decoder.Decode(&obj.TokenBMint)
	if err != nil {
		return err
	}
	// Deserialize `TokenAAccount`:
	err = decoder.Decode(&obj.TokenAAccount)
	if err != nil {
		return err
	}
	// Deserialize `TokenBAccount`:
	err = decoder.Decode(&obj.TokenBAccount)
	if err != nil {
		return err
	}
	// Deserialize `CurrentSqrtPrice`:
	err = decoder.Decode(&obj.CurrentSqrtPrice)
	if err != nil {
		return err
	}
	// Deserialize `IsPause`:
	err = decoder.Decode(&obj.IsPause)
	if err != nil {
		return err
	}
	// Deserialize `PositionNum`:
	err = decoder.Decode(&obj.PositionNum)
	if err != nil {
		return err
	}
	// Deserialize `IsCrmReward`:
	err = decoder.Decode(&obj.IsCrmReward)
	if err != nil {
		return err
	}
	return nil
}