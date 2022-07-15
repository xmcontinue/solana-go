// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package refundPosition

import ag_binary "github.com/gagliardetto/binary"

type ErrorCode ag_binary.BorshEnum

const (
	ErrorCodeCremaPositionNotFound ErrorCode = iota
	ErrorCodeInvalidSwapAccount
	ErrorCodeSwapMintError
	ErrorCodeSwapPaused
	ErrorCodePositionAlreadyRefunded
	ErrorCodeSwapRewardStatusError
)

func (value ErrorCode) String() string {
	switch value {
	case ErrorCodeCremaPositionNotFound:
		return "CremaPositionNotFound"
	case ErrorCodeInvalidSwapAccount:
		return "InvalidSwapAccount"
	case ErrorCodeSwapMintError:
		return "SwapMintError"
	case ErrorCodeSwapPaused:
		return "SwapPaused"
	case ErrorCodePositionAlreadyRefunded:
		return "PositionAlreadyRefunded"
	case ErrorCodeSwapRewardStatusError:
		return "SwapRewardStatusError"
	default:
		return ""
	}
}