package client

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/pkg/iface"
)

type ExchangeClient struct {
	*rpcx.Client
}

func (c *ExchangeClient) GetPrice(ctx context.Context, args *iface.GetPriceReq, reply *iface.GetPriceResp) error {
	return c.Call(ctx, "GetPrice", args, reply)
}

// NewCremaExchangeClient NewCremaExchangeClient Rpc客户端
func NewCremaExchangeClient(ctx context.Context, config *transport.ServiceConfig) (iface.ExchangeService, error) {
	client, err := rpcx.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &ExchangeClient{client}, nil
}
