package worker

import (
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
)

type swapPairPrice struct {
	TokenASymbol string
	TokenBSymbol string
	Price        decimal.Decimal
}

type tokenPrice struct {
	Price decimal.Decimal
}

func SyncSwapPrice() error {
	logger.Info("price syncing ......")

	return nil
}
