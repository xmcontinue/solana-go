package handler

import (
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	AddLog = handleFunc(exporterClient, "AddLog", &iface.LogReq{}, &iface.LogResp{})
)
