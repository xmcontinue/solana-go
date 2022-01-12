package handler

import (
	"context"

	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/common/chain/sol"
	"git.cplus.link/crema/backend/common/pkg/iface"
)

// SwapCount ...
func (t *ToolService) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)
	reply.List = sol.GetCount(args)
	return nil
}
