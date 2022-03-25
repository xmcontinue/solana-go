package exchanger

import (
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/internal/market"
)

type Exchanger struct {
	Market *market.Market
	data   *Data
}

func NewExchanger() *Exchanger {
	e := &Exchanger{
		market.NewMarket(),
		NewData(),
	}
	return e
}

func (e *Exchanger) LoadConfig(eConfig *config.ExchangeConfig) error {
	err := e.Market.LoadConfig(eConfig)
	if err != nil {
		return err
	}
	return nil
}

func (e *Exchanger) SyncPrice() error {
	raw, err := e.Market.GetPrices()
	if err != nil {
		return errors.Wrap(err)
	}

	e.data.LoadRawData(raw)

	return nil
}

func (e *Exchanger) GetPriceForMarketForShotPath(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	return e.data.GetPriceForMarketForShotPath(name, strings.ToUpper(baseSymbol), strings.ToUpper(quoteSymbol))
}

func (e *Exchanger) GetPricesForMarket(name, quoteSymbol string) (*Prices, error) {
	return e.data.GetPricesForMarket(name, strings.ToUpper(quoteSymbol))
}

func (e *Exchanger) GetPriceForMarket(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	return e.data.GetPriceForMarket(name, strings.ToUpper(baseSymbol), strings.ToUpper(quoteSymbol))
}
