package handler

import (
	"git.cplus.link/crema/backend/common/pkg/iface"
)

var (
	SwapCount = handleFunc(toolClient, "SwapCount", &iface.SwapCountReq{}, &iface.SwapCountResp{})
)
