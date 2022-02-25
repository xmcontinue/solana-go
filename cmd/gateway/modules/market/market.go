package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	SwapCount   = handleFunc(marketClient, "SwapCount", &iface.SwapCountReq{}, &iface.SwapCountResp{})
	GetConfig   = handleFunc(marketClient, "GetConfig", &iface.GetConfigReq{}, &iface.JsonString{})
	GetTvl      = handleFunc(marketClient, "GetTvl", &iface.GetTvlReq{}, &iface.GetTvlResp{})
	GetTvlV2    = handleFunc(marketClient, "GetTvlV2", &iface.GetTvlReqV2{}, &iface.GetTvlRespV2{})
	Get24hVolV2 = handleFunc(marketClient, "Get24hVolV2", &iface.Get24hVolV2Req{}, &iface.Get24hVolV2Resp{})
	GetVolV2    = handleFunc(marketClient, "GetVolV2", &iface.GetVolV2Req{}, &iface.GetVolV2Resp{})
)
