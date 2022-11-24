package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func clearSwapCountKline() error {
	swapConfigs := sol.SwapConfigList()

	beforeOneYear := time.Now().UTC().Add(-time.Hour * 24 * 30 * 12)
	for _, v := range swapConfigs {
		err := model.DeleteSwapCountKLines(
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
