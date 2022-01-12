package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	watcher "git.cplus.link/crema/backend/common/internal/worker/market"

	"git.cplus.link/crema/backend/common/pkg/iface"
)

// SwapCount ...
func (t *MarketService) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)

	reply.SwapPairCount = watcher.GetSwapCountCache(args.TokenSwapAddress)

	if reply.SwapPairCount == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
