package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	GetPrices = handleFunc(exchangeClient, "GetPrice", &iface.GetPriceReq{}, &iface.GetPriceResp{})
)
