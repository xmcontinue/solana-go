package market

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	tvlCache *domain.Tvl
)

// TvlCacheJob ...
func TvlCacheJob() error {
	logger.Info("tvl cache setting ......")

	tvl, err := model.QueryTvl(context.TODO())
	if err != nil {
		logger.Info("tvl cache set fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}

	tvlCache = tvl

	logger.Info("tvl cache set complete!")
	return nil
}

// GetTvlCache ....
func GetTvlCache() *domain.Tvl {
	return tvlCache
}
