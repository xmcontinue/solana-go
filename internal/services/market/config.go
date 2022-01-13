package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/worker/market"

	"git.cplus.link/crema/backend/pkg/iface"
)

// GetConfig ...
func (t *MarketService) GetConfig(ctx context.Context, args *iface.GetConfigReq, reply *iface.GetConfigResp) error {
	defer rpcx.Recover(ctx)

	reply.Data = market.GetConfig(args.ConfigName)

	if reply.Data == nil {
		reply.Data = []byte("{}")
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
