package parse

import (
	"git.cplus.link/go/akit/util/decimal"

	event "git.cplus.link/crema/backend/chain/event/parser"
	"git.cplus.link/crema/backend/pkg/domain"
)

type CollectRecordV2 struct {
	SwapConfig       *domain.SwapConfig
	UserOwnerAddress string
	AmountA          uint64
	AmountB          uint64
}

type CollectRecordIface interface {
	GetSwapConfig() *domain.SwapConfig
	GetUserOwnerAccount() string
	GetTokenACollectVolume() decimal.Decimal
	GetTokenBCollectVolume() decimal.Decimal
}

func (c *CollectRecordV2) GetSwapConfig() *domain.SwapConfig {
	return c.SwapConfig
}

func (c *CollectRecordV2) GetUserOwnerAccount() string {
	return c.UserOwnerAddress
}

func (c *CollectRecordV2) GetTokenACollectVolume() decimal.Decimal {
	return PrecisionConversion(decimal.NewFromInt(int64(c.AmountA)), int(c.GetSwapConfig().TokenA.Decimal))
}

func (c *CollectRecordV2) GetTokenBCollectVolume() decimal.Decimal {
	return PrecisionConversion(decimal.NewFromInt(int64(c.AmountA)), int(c.GetSwapConfig().TokenB.Decimal))
}

func (t *Txv2) createCollectRecord(logMessageEvent event.EventRep) error {
	collectEvent := logMessageEvent.Event.(*event.CollectEvent)

	swapConfig, ok := swapConfigMap[collectEvent.Pool.String()]
	if !ok {
		return nil
	}

	t.ClaimRecords = append(t.ClaimRecords, &CollectRecordV2{
		SwapConfig:       swapConfig,
		UserOwnerAddress: collectEvent.Owner.String(),
		AmountA:          collectEvent.AmountA,
		AmountB:          collectEvent.AmountB,
	})

	return nil
}
