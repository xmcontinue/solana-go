package market

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/internal/market/bybit"
	"git.cplus.link/crema/backend/internal/market/coingecko"
	"git.cplus.link/crema/backend/internal/market/ftx"
	"git.cplus.link/crema/backend/internal/market/gate"
	"git.cplus.link/crema/backend/pkg/domain"
)

type Market struct {
	Businesses []Business
	config     *config.ExchangeConfig
}

func NewMarket() *Market {
	m := &Market{}
	return m
}

func (m *Market) LoadConfig(eConfig *config.ExchangeConfig) error {
	businesses := make([]Business, 0)

	names := []string{
		coingecko.BusinessName,
		ftx.BusinessName,
	}

	for _, name := range names {
		switch name {
		case coingecko.BusinessName:
			coinGeckoS, err := coingecko.NewCoinGecko(eConfig)
			if err != nil {
				return errors.Wrap(err)
			}
			businesses = append(businesses, coinGeckoS)
		case bybit.BusinessName:
			byBitS, err := bybit.NewByBit(eConfig)
			if err != nil {
				return errors.Wrap(err)
			}
			businesses = append(businesses, byBitS)
		case gate.BusinessName:
			businesses = append(businesses, gate.NewGate(eConfig))
		case ftx.BusinessName:
			businesses = append(businesses, ftx.NewFtx(eConfig))
		default:
			continue
		}
	}

	m.Businesses = businesses
	m.config = eConfig

	return nil
}

func (m *Market) GetPrices() (map[string]map[string][]*domain.Price, error) {
	if len(m.Businesses) == 0 {
		return nil, errors.New("business is not found")
	}

	prices := make(map[string]map[string][]*domain.Price)

	for _, v := range m.Businesses {
		l, err := v.GetPrices()
		if err != nil {
			continue
		}

		l = m.ReplaceSymbolsPrice(l)

		prices[v.GetName()] = l
	}

	return prices, nil
}

func (m *Market) ReplaceSymbolsPrice(coins map[string][]*domain.Price) map[string][]*domain.Price {
	for k, prices := range coins {
		replacePrices := make(map[string]decimal.Decimal, 0)
		for _, v := range m.GetConfig().GetReplaceSymbols() {
			replacePrices[v] = decimal.Decimal{}
		}

		for _, price := range prices {
			if _, ok := replacePrices[price.BaseSymbol]; ok {
				replacePrices[price.BaseSymbol] = price.Price
			}
		}

		for i, v := range m.GetConfig().GetReplaceSymbols() {
			if price, ok := replacePrices[v]; ok {
				if !price.IsZero() {
					coins[k] = append(coins[k], &domain.Price{
						BaseSymbol:  i,
						QuoteSymbol: k,
						Price:       price,
					})
				}
			}
		}
	}

	return coins
}

func (m *Market) GetConfig() *config.ExchangeConfig {
	return m.config
}

func (m *Market) setConfig(eConfig *config.ExchangeConfig) error {
	m.config = eConfig
	return nil
}

type Business interface {
	GetName() string
	GetPrices() (map[string][]*domain.Price, error)
}
