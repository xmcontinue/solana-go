package exchanger

import (
	"strings"
	"time"

	aConfig "git.cplus.link/go/akit/config"
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

	e.data.LoadConfig(eConfig)
	return nil
}

func (e *Exchanger) SyncPrice() error {
	raw, err := e.Market.GetPrices()
	if err != nil {
		return errors.Wrap(err)
	}

	e.data.DataHandle(raw)

	e.data.LoadRawData(raw)

	e.data.LoadAvgData()

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

func (e *Exchanger) RunForViper(aConf *aConfig.Config) error {
	if err := e.watchConfigForViper(aConf); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (e *Exchanger) watchConfigForViper(aConf *aConfig.Config) error {
	err := aConf.WatchRemoteConfigOnChannel()
	if err != nil {
		return errors.Wrap(err)
	}

	var initEConfigF = func() error {
		exchangeConf, err := config.NewExchangeConfigForViper(aConf)
		if err != nil {
			return errors.Wrap(err)
		}

		if e.Market.GetConfig() != nil && e.Market.GetConfig().Equal(exchangeConf) {
			return nil
		}

		err = e.LoadConfig(exchangeConf)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}

	if err = initEConfigF(); err != nil {
		return errors.Wrap(err)
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)

			_ = initEConfigF()
		}
	}()

	err = e.SyncPrice()
	if err != nil {
		return err
	}

	return nil
}
