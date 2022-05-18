package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	SwapCount                = handleFunc(marketClient, "SwapCount", &iface.NilReq{}, &iface.SwapCountResp{})
	SwapCountList            = handleFunc(marketClient, "SwapCountList", &iface.SwapCountListReq{}, &iface.SwapCountListResp{})
	SwapCountNew             = handleFunc(marketClient, "SwapCountNew", &iface.NilReq{}, &iface.SwapCountResp{})
	GetConfig                = handleFunc(marketClient, "GetConfig", &iface.GetConfigReq{}, &iface.JsonString{})
	GetTokenConfig           = handleFunc(marketClient, "GetTokenConfig", &iface.NilReq{}, &iface.JsonString{})
	GetTvl                   = handleFunc(marketClient, "GetTvl", &iface.GetTvlReq{}, &iface.GetTvlResp{})
	GetTvlV2                 = handleFunc(marketClient, "GetTvlV2", &iface.GetTvlReqV2{}, &iface.GetTvlRespV2{})
	Get24hVolV2              = handleFunc(marketClient, "Get24hVolV2", &iface.Get24hVolV2Req{}, &iface.Get24hVolV2Resp{})
	GetVolV2                 = handleFunc(marketClient, "GetVolV2", &iface.GetVolV2Req{}, &iface.GetVolV2Resp{})
	GetKline                 = handleFunc(marketClient, "GetKline", &iface.GetKlineReq{}, &iface.GetKlineResp{})
	GetHistogram             = handleFunc(marketClient, "GetHistogram", &iface.GetHistogramReq{}, &iface.GetHistogramResp{})
	TvlOfSingleToken         = handleFunc(marketClient, "TvlOfSingleToken", &iface.TvlOfSingleTokenReq{}, &iface.TvlOfSingleTokenResp{})
	GetActivityHistoryByUser = handleFunc(marketClient, "GetActivityHistoryByUser", &iface.GetActivityHistoryByUserReq{}, &iface.GetActivityHistoryByUserResp{})
	GetActivityNftMetadata   = handleFuncForNft(marketClient, "GetActivityNftMetadata", &iface.GetActivityNftMetadataReq{}, &iface.GetActivityNftMetadataResp{})
)
