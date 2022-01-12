package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	wacher "git.cplus.link/crema/backend/common/internal/worker/tool"
	"git.cplus.link/crema/backend/common/pkg/iface"
)

// SwapCount ...
func (t *ToolService) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)

	reply.TokenVolumeCount = wacher.GetSwapCountCache(args.TokenSwapAddress)

	if reply.TokenVolumeCount == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
