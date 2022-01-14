package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/worker/market"

	"git.cplus.link/crema/backend/pkg/iface"
)

// GetConfig ...
func (t *MarketService) GetConfig(ctx context.Context, args *iface.GetConfigReq, reply *iface.JsonString) error {
	defer rpcx.Recover(ctx)

	*reply = market.GetConfig(args.Name)

	if reply == nil {
		*reply = []byte("{}")
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
