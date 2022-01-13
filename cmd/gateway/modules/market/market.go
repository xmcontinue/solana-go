package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	SwapCount = handleFunc(marketClient, "SwapCount", &iface.SwapCountReq{}, &iface.SwapCountResp{})
)
