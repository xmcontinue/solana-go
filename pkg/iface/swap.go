package iface

import (
	"context"

	"git.cplus.link/crema/backend/pkg/domain"
)

const MarketServiceName = "CremaMarketService"

type MarketService interface {
	SwapCount(context.Context, *SwapCountReq, *SwapCountResp) error
	GetConfig(context.Context, *GetConfigReq, *JsonString) error
	GetTvl(context.Context, *GetTvlReq, *GetTvlResp) error
}

type SwapCountReq struct {
	TokenSwapAddress string `form:"token_swap_address"      binding:"required"`
}

type SwapCountResp struct {
	*domain.SwapPairCount `json:"swap_pair_count"`
}

type GetConfigReq struct {
	Name string `form:"name"      binding:"required"`
}

type JsonString []byte

func (j *JsonString) MarshalJSON() ([]byte, error) {
	return *j, nil
}

func (j *JsonString) UnmarshalJSON(data []byte) error {
	*j = data
	return nil
}

type GetTvlReq struct {
}

type GetTvlResp struct {
	*domain.Tvl
}
