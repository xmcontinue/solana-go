package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/pkg/iface"
	"git.cplus.link/crema/backend/pkg/prometheus"
)

// AddLog 添加log至prometheus 的 push gateway
func (t *ExporterService) AddLog(ctx context.Context, args *iface.LogReq, _ *iface.LogResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	if len(args.Tags) > 10 {
		return errors.New("tags too long")
	}
	if _, ok := args.Tags["project"]; !ok {
		args.Tags["project"] = prometheus.GetProjectName()
	}

	if err := prometheus.CheckAuth(args); err != nil {
		return errors.Wrap(err)
	}

	if err := prometheus.ExamplePusherPush(args); err != nil {
		return errors.Wrapf(errors.NotReady, "logging failed:%v", err)
	}
	return nil
}
