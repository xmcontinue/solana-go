package handler

import (
	"git.cplus.link/crema/backend/common/pkg/iface"
)

var (
	SwapCount = handleFunc(marketClient, "SwapCount", &iface.SwapCountReq{}, &iface.SwapCountResp{})
)
