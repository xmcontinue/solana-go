package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/worker/market"

	"git.cplus.link/crema/backend/pkg/iface"
)

// SwapCount ...
func (t *MarketService) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)

	reply.SwapPairCount = market.GetSwapCountCache(args.TokenSwapAddress)

	if reply.SwapPairCount == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}

// GetTvl ...
func (t *MarketService) GetTvl(ctx context.Context, args *iface.GetTvlReq, reply *iface.GetTvlResp) error {
	defer rpcx.Recover(ctx)

	reply.Tvl = market.GetTvlCache()

	if reply.Tvl == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
