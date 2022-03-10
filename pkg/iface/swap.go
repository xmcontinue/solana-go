package iface

import (
	"context"

	"git.cplus.link/go/akit/util/decimal"
	"git.cplus.link/go/akit/util/gquery"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/internal/worker/process"
	"git.cplus.link/crema/backend/pkg/domain"
)

const MarketServiceName = "CremaMarketService"

type MarketService interface {
	GetConfig(context.Context, *GetConfigReq, *JsonString) error
	GetTvl(context.Context, *GetTvlReq, *GetTvlResp) error
	SwapCount(context.Context, *NilReq, *SwapCountResp) error
}

type SwapCountReq struct {
	TokenSwapAddress string `form:"token_swap_address"      binding:"required"`
}

type SwapCountResp struct {
	*domain.SwapCountToApi
}

type SwapCountOldResp struct {
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

type NilReq struct {
}
type GetTvlReq struct {
}

type GetTvlResp struct {
	*domain.Tvl
}

type GetTvlReqV2 struct {
	SwapAddress string `json:"swap_address"   binding:"omitempty"` // 如果不为空，表示查询当前的swap address 的tvl ,如果为空，表示查询配置了的swap address
}

type SwapAddressTvl struct {
	SwapAccount string
	Tvl         decimal.Decimal
}
type GetTvlRespV2 struct {
	List []*SwapAddressTvl
}

type Get24hVolV2Req struct {
	AccountAddress string `json:"account_address"           binding:"required"`
}

type Get24hVolV2Resp struct {
	*model.SwapVol
}

type GetVolV2Req struct {
	SwapAddress string `json:"swap_address"           binding:"required"`
}

type GetVolV2Resp struct {
	Vol decimal.Decimal `json:"vol"`
}

type GetNetRecordReq struct {
	Limit   int              `json:"limit,omitempty"        form:"limit"        gquery:"-"`                        // limit
	Offset  int              `json:"offset,omitempty"       form:"offset"       gquery:"-"`                        // offset
	OrderBy *gquery.GOrderBy `json:"order_by,omitempty"     form:"order_by"     gquery:"id,updated_at,created_at"` // 排序
}

type GetNetRecordResp struct {
	Total  int64               `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
	List   []*domain.NetRecode `json:"list"`
}

type QueryUserSwapTvlCountReq struct {
	UserAddress string           `json:"user_address" binding:"omitempty"`
	Limit       int              `json:"limit,omitempty"        form:"limit"        gquery:"-"`                        // limit
	Offset      int              `json:"offset,omitempty"       form:"offset"       gquery:"-"`                        // offset
	OrderBy     *gquery.GOrderBy `json:"order_by,omitempty"     form:"order_by"     gquery:"id,updated_at,created_at"` // 排序
}

type QueryUserSwapTvlCountResp struct {
	Total  int64                   `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
	List   []*domain.UserSwapCount `json:"list"`
}

type QueryUserSwapTvlCountDayReq struct {
	UserAddress string           `json:"user_address" binding:"omitempty"`
	Limit       int              `json:"limit,omitempty"        form:"limit"        gquery:"-"`                        // limit
	Offset      int              `json:"offset,omitempty"       form:"offset"       gquery:"-"`                        // offset
	OrderBy     *gquery.GOrderBy `json:"order_by,omitempty"     form:"order_by"     gquery:"id,updated_at,created_at"` // 排序
}

type QueryUserSwapTvlCountDayResp struct {
	Total  int64                      `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
	List   []*domain.UserSwapCountDay `json:"list"`
}

type GetKlineReq struct {
	SwapAccount string          `json:"swap_account"      binding:"required"`
	DateType    domain.DateType `json:"date_type"         binding:"required"`
	Limit       int             `json:"limit,omitempty"        form:"limit"        gquery:"-"` // limit
	Offset      int             `json:"offset,omitempty"       form:"offset"       gquery:"-"` // offset
}

type GetKlineResp struct {
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	List   []*process.Price `json:"list"`
}
type GetHistogramReq struct {
	Typ         string          `form:"typ"               binding:"required,oneof=tvl vol"`
	SwapAccount string          `form:"swap_account"      binding:"omitempty"`
	DateType    domain.DateType `form:"date_type"         binding:"required,oneof=1min 5min 15min 30min hour day wek mon "`
	Limit       int             `form:"limit,omitempty"        form:"limit"        gquery:"-"` // limit
	Offset      int             `form:"offset,omitempty"       form:"offset"       gquery:"-"` // offset
}

type GetHistogramResp struct {
	Total  int64                         `json:"total"`
	Limit  int                           `json:"limit"`
	Offset int                           `json:"offset"`
	List   []*process.SwapHistogramPrice `json:"list"`
}