package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type ParserTransaction interface {
	GetSyncPoint() error
	WriteToDB(*domain.SwapTransaction) error
	ParserDate() error
}

func parserUserCountAndSwapCount() error {

	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}

	for _, swapConfig := range sol.SwapConfigList() {

		swapPairBase, err := model.QuerySwapPairBase(context.TODO(), model.SwapAddress(swapConfig.SwapAccount))
		if err != nil {
			logger.Error("query swap_pair_bases err", logger.Errorv(err))
			return errors.Wrap(err)
		}
		if swapPairBase == nil {
			break
		}

		if swapPairBase.IsSync == false {
			break
		}

		swapAndUserCount := &SwapAndUserCount{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       swapConfig.SwapAccount,
		}

		if err = swapAndUserCount.GetSyncPoint(); err != nil {
			return errors.Wrap(err)
		}

		if err = swapAndUserCount.ParserDate(); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}
