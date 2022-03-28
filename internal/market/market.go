package market

import (
	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/internal/market/coingecko"
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
	name := coingecko.BusinessName
	switch name {
	case coingecko.BusinessName:
		coingeckoS, err := coingecko.NewCoingecko(eConfig)
		if err != nil {
			return errors.Wrap(err)
		}
		businesses = append(businesses, coingeckoS)
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

		prices[v.GetName()] = l
	}

	return prices, nil
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
