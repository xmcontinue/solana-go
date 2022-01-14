package market

import (
	"context"

	"git.cplus.link/go/akit/errors"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	tvlCache *domain.Tvl
)

// TvlCacheJob ...
func TvlCacheJob() error {

	tvl, err := model.QueryTvl(context.TODO())
	if err != nil {
		return errors.Wrap(err)
	}

	tvlCache = tvl

	return nil
}

// GetTvlCache ....
func GetTvlCache() *domain.Tvl {
	return tvlCache
}
