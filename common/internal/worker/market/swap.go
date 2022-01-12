package watcher

import (
	"context"

	model "git.cplus.link/crema/backend/common/internal/model/market"

	"git.cplus.link/crema/backend/common/chain/sol"
	"git.cplus.link/crema/backend/common/pkg/domain"
)

var (
	swapPairCountCache map[string]*domain.SwapPairCount
)

// SwapCountCacheJob ...
func SwapCountCacheJob() error {

	swapPairCountMap := make(map[string]*domain.SwapPairCount)

	for _, v := range sol.Addresses() {

		count, err := model.QuerySwapPairCount(context.Background(), model.NewFilter("token_swap_address = ?", v.TokenSwapAddress))

		if err != nil {
			continue
		}

		swapPairCountMap[v.TokenSwapAddress] = count

	}

	swapPairCountCache = swapPairCountMap

	return nil
}

// GetSwapCountCache ...
func GetSwapCountCache(tokenSwapAddress string) *domain.SwapPairCount {
	return swapPairCountCache[tokenSwapAddress]
}
