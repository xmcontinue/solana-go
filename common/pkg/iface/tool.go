package iface

import (
	"context"

	"git.cplus.link/crema/backend/common/pkg/domain"
)

const ToolServiceName = "CremaToolService"

type ToolService interface {
	SwapCount(context.Context, *SwapCountReq, *SwapCountResp) error
}

type SwapCountReq struct {
	TokenSwapAddress string `json:"token_swap_address"      binding:"required"`
}

type SwapCountResp struct {
	*domain.SwapPairCount `json:"swap_pair_count"`
}
