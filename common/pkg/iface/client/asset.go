package client

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/common/pkg/iface"
)

type CremaToolClient struct {
	*rpcx.Client
}

func (c *CremaToolClient) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	return c.Call(ctx, "SwapCount", args, reply)
}

// NewCremaToolClient NewCremaToolClient Rpc客户端
func NewCremaToolClient(ctx context.Context, config *transport.ServiceConfig) (iface.ToolService, error) {
	client, err := rpcx.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &CremaToolClient{client}, nil
}
