package process

import (
	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
)

type ParserTransaction interface {
	GetBeginTransactionID() error
	WriteToDB(*domain.SwapTransaction) error
	ParserDate() error
}

type syncType string

type SwapTokenIndex struct {
	TokenAIndex int64
	TokenBIndex int64
}

var (
	swapAccount = "8J3avAjuRfL2CYFKKDwhhceiRoajhrHv9kN5nUiEnuBG"
)

func parserUserCountAndSwapCount() error {

	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}

	for _, swapConfig := range sol.SwapConfigList() {
		swapAndUserCount := &SwapAndUserCount{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       swapConfig.SwapAccount,
		}

		if err = swapAndUserCount.GetBeginTransactionID(); err != nil {
			return errors.Wrap(err)
		}

		if err = swapAndUserCount.ParserDate(); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}
