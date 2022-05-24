package client

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/pkg/iface"
)

type CremaExporterClient struct {
	*rpcx.Client
}

func (c *CremaExporterClient) AddLog(ctx context.Context, args *iface.LogReq, reply *iface.LogResp) error {
	return c.Call(ctx, "SwapCount", args, reply)
}

// NewCremaExporterClient Rpc客户端
func NewCremaExporterClient(ctx context.Context, config *transport.ServiceConfig) (iface.ExporterService, error) {
	client, err := rpcx.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &CremaExporterClient{client}, nil
}
