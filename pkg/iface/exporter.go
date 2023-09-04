package iface

import (
	"context"
)

const ExporterServiceName = "CremaExporterService"

type ExporterService interface {
	AddLog(context.Context, *LogReq, *LogResp) error
}

type LogReq struct {
	Key      string            `json:"key" binding:"required"`
	LogName  string            `json:"log_name" binding:"required"`
	LogValue float64           `json:"log_value" binding:"required"`
	LogHelp  string            `json:"log_help" binding:"required"`
	JobName  string            `json:"job_name" binding:"required"`
	Tags     map[string]string `json:"tags" binding:"required"`
}

type LogResp struct {
}
