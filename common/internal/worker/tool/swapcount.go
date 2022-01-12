package wacher

import (
	"context"

	"git.cplus.link/crema/backend/common/chain/sol"
	model "git.cplus.link/crema/backend/common/internal/model/tool"
	"git.cplus.link/crema/backend/common/pkg/domain"
)

var (
	swapCountCache map[string]*domain.TokenVolumeCount
)

// SwapCountCacheJob ...
func SwapCountCacheJob() error {

	SwapCountMap := make(map[string]*domain.TokenVolumeCount)

	for _, v := range sol.Addresses() {

		count, err := model.QueryTokenVolumeCount(context.Background(), model.NewFilter("token_swap_address = ?", v.TokenSwapAddress))

		if err != nil {
			continue
		}

		SwapCountMap[v.TokenSwapAddress] = count

	}

	swapCountCache = SwapCountMap

	return nil
}

// GetSwapCountCache ...
func GetSwapCountCache(tokenSwapAddress string) *domain.TokenVolumeCount {
	return swapCountCache[tokenSwapAddress]
}
