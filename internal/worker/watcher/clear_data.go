package watcher

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func clearTokenPriceKline() error {
	beforeOneMonth := time.Now().UTC().Add(-time.Hour * 24 * 30)
	err := model.DeleteSwapTokenPriceKLine(context.Background(),
		model.NewFilter("updated_at < ?", beforeOneMonth),
		model.DateTypeFilter(domain.DateMin),
	)

	return errors.Wrap(err)
}

func clearSwapTransactions() error {
	swapConfigs := sol.SwapConfigList()

	beforeOneMonth := time.Now().UTC().Add(-time.Hour * 24 * 30)
	for _, v := range swapConfigs {
		err := model.DeleteSwapTransactionV2(
			context.Background(),
			model.SwapAddressFilter(v.SwapAccount),
			model.NewFilter("block_time < ?", beforeOneMonth),
			model.DateTypeFilter(domain.DateMin),
		)
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			return errors.Wrap(err)
		}
	}

	return nil
}

func clearSwapPriceKline() error {
	swapConfigs := sol.SwapConfigList()

	beforeOneYear := time.Now().UTC().Add(-time.Hour * 24 * 30 * 12)
	for _, v := range swapConfigs {
		err := model.DeleteSwapPairPriceKLine(
			context.Background(),
			model.SwapAddressFilter(v.SwapAccount),
			model.NewFilter("date < ?", beforeOneYear),
			model.DateTypeFilter(domain.DateMin),
		)
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			return errors.Wrap(err)
		}
	}

	return nil
}

func ClearOldData() error {

	// 清除一个月以前的数据
	err := clearSwapTransactions()
	if err != nil {
		return errors.Wrap(err)
	}

	err = clearSwapPriceKline()
	if err != nil {
		return errors.Wrap(err)
	}

	// 清除一个月以前的颗粒度细的数据
	err = clearTokenPriceKline()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
