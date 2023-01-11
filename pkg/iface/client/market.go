package client

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/pkg/iface"
)

type CremaMarketClient struct {
	*rpcx.Client
}

func (c *CremaMarketClient) SwapCount(ctx context.Context, args *iface.NilReq, reply *iface.SwapCountResp) error {
	return c.Call(ctx, "SwapCount", args, reply)
}

func (c *CremaMarketClient) SwapCountList(ctx context.Context, args *iface.SwapCountListReq, reply *iface.SwapCountListResp) error {
	return c.Call(ctx, "SwapCountList", args, reply)
}

func (c *CremaMarketClient) GetConfig(ctx context.Context, args *iface.GetConfigReq, reply *iface.JsonString) error {
	return c.Call(ctx, "GetConfig", args, reply)
}
func (c *CremaMarketClient) GetTokenConfig(ctx context.Context, args *iface.NilReq, reply *iface.JsonString) error {
	return c.Call(ctx, "GetTokenConfig", args, reply)
}
func (c *CremaMarketClient) GetTvl(ctx context.Context, args *iface.GetTvlReq, reply *iface.GetTvlResp) error {
	return c.Call(ctx, "GetTvl", args, reply)
}

func (c *CremaMarketClient) GetKline(ctx context.Context, args *iface.GetKlineReq, reply *iface.GetKlineResp) error {
	return c.Call(ctx, "GetKline", args, reply)
}

func (c *CremaMarketClient) GetHistogram(ctx context.Context, args *iface.GetHistogramReq, reply *iface.GetHistogramResp) error {
	return c.Call(ctx, "GetHistogram", args, reply)
}

func (c *CremaMarketClient) TvlOfSingleToken(ctx context.Context, args *iface.TvlOfSingleTokenReq, reply *iface.TvlOfSingleTokenResp) error {
	return c.Call(ctx, "TvlOfSingleToken", args, reply)
}

func (c *CremaMarketClient) GetTransactions(ctx context.Context, args *iface.GetTransactionsReq, reply *iface.GetTransactionsResp) error {
	return c.Call(ctx, "GetTransactions", args, reply)
}

func (c *CremaMarketClient) QuerySwapKline(ctx context.Context, args *iface.QuerySwapKlineReq, reply *iface.QuerySwapKlineResp) error {
	return c.Call(ctx, "QuerySwapKline", args, reply)
}

func (c *CremaMarketClient) QueryPositions(ctx context.Context, args *iface.QueryPositionsReq, reply *iface.QueryPositionsResp) error {
	return c.Call(ctx, "QueryPositions", args, reply)
}

func (c *CremaMarketClient) GetGallery(ctx context.Context, args *iface.GetGalleryReq, reply *iface.GetGalleryResp) error {
	return c.Call(ctx, "GetGallery", args, reply)
}

func (c *CremaMarketClient) QueryPriceForSymbol(ctx context.Context, args *iface.QueryPriceForSymbolReq, reply *iface.QueryPriceForSymbolResp) error {
	return c.Call(ctx, "QueryPriceForSymbol", args, reply)
}

// NewCremaMarketClient NewCremaMarketClient Rpc客户端
func NewCremaMarketClient(ctx context.Context, config *transport.ServiceConfig) (iface.MarketService, error) {
	client, err := rpcx.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &CremaMarketClient{client}, nil
}
