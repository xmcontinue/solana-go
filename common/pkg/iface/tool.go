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
	TokenAPoolAddress string `json:"token_a_pool_address"`
	TokenBPoolAddress string `json:"token_b_pool_address"`
	TokenSwapAddress  string `json:"token_swap_address"`
}

type SwapCountResp struct {
	List []*domain.TokenVolumeCount `json:"list"`
}
