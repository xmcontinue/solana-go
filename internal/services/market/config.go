package handler

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/worker/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

// GetConfig ...
func (t *MarketService) GetConfig(ctx context.Context, args *iface.GetConfigReq, reply *iface.JsonString) error {
	defer rpcx.Recover(ctx)

	*reply = market.GetConfig(args.Name)

	if *reply == nil {
		*reply = []byte("{}")
		return errors.Wrap(errors.RecordNotFound)
	}

	if args.Name == "v2-swap-pairs" {
		swapConfigListV2 := &[]*domain.SwapConfig{}

		err := json.Unmarshal(*reply, swapConfigListV2)
		if err != nil {
			logger.Error("swap config unmarshal failed :", logger.Errorv(err))
		}

		statusByte, err := t.redisClient.Get(ctx, domain.SwapStatusKey().Key).Result()
		if err != nil {
			return errors.Wrap(err)
		}
		statusM := make(map[string]bool)
		_ = json.Unmarshal([]byte(statusByte), &statusM)

		for _, v := range *swapConfigListV2 {
			status, ok := statusM[v.SwapAccount]
			if !ok {
				return errors.Wrap(errors.RecordNotFound)
			}
			v.IsPause = status
		}
	}
	return nil
}

// GetTokenConfig ...
func (t *MarketService) GetTokenConfig(ctx context.Context, _ *iface.NilReq, reply *iface.JsonString) error {
	defer rpcx.Recover(ctx)
	*reply = market.GetTokenConfig()

	if *reply == nil {
		*reply = []byte("{}")
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}
