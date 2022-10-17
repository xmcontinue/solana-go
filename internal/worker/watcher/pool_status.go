package watcher

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/pkg/domain"
)

func syncSwapStatus() error {

	configs := sol.SwapConfigListV2()

	swapStatus := make(map[string]bool)
	// 同步swap pair price
	for _, config := range configs {
		res, err := sol.GetRpcClient().GetAccountInfo(context.Background(), config.SwapPublicKey)
		if err != nil {
			logger.Error("get swap status err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		isPause, err := parse.GetPoolPauseStatus(res)
		if err != nil {
			logger.Error("parse swap status err", logger.Errorv(err))
			return errors.Wrap(err)
		}
		swapStatus[config.SwapAccount] = isPause
	}

	data, _ := json.Marshal(swapStatus)
	err := redisClient.Set(context.Background(), domain.SwapStatusKey().Key, data, domain.SwapStatusKey().Timeout).Err()
	if err != nil {
		logger.Error("set swap status to redis err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil
}
