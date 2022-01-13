package client

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/common/pkg/iface"
)

type CremaMarketClient struct {
	*rpcx.Client
}

func (c *CremaMarketClient) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	return c.Call(ctx, "SwapCount", args, reply)
}

func (c *CremaMarketClient) GetConfig(ctx context.Context, args *iface.GetConfigReq, reply *iface.GetConfigResp) error {
	return c.Call(ctx, "GetConfig", args, reply)
}

// NewCremaMarketClient NewCremaMarketClient Rpc客户端
func NewCremaMarketClient(ctx context.Context, config *transport.ServiceConfig) (iface.MarketService, error) {
	client, err := rpcx.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &CremaMarketClient{client}, nil
}
