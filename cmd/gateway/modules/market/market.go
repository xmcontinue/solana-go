package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	SwapCount = handleFunc(marketClient, "SwapCount", &iface.SwapCountReq{}, &iface.SwapCountResp{})
	GetConfig = handleFunc(marketClient, "GetConfig", &iface.GetConfigReq{}, &iface.JsonString{})
	GetTvl    = handleFunc(marketClient, "GetTvl", &iface.GetTvlReq{}, &iface.GetTvlResp{})
)
