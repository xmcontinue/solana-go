package exchanger

import (
	"fmt"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/exchanger/graph"
	"git.cplus.link/crema/backend/pkg/domain"
)

type Data struct {
	raw Markets
	avg Coins
}

func NewData() *Data {
	return &Data{}
}

func (d *Data) LoadRawData(raw map[string]map[string][]*domain.Price) {
	r := make(Markets, len(raw))
	for p, qs := range raw {
		quote := make(Coins, len(qs))
		for s, v := range qs {
			quote[s] = v
		}
		r[p] = quote
	}
	d.raw = r
}

func (d *Data) GetPriceForMarketForShotPath(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	var price decimal.Decimal

	coins, ok := d.raw[name]
	if !ok {
		return price, errors.New("market is not found")
	}

	price, err := coins.GetPriceForShotPath(baseSymbol, quoteSymbol)
	if err != nil {
		return price, errors.Wrap(err)
	}

	return price, nil
}

func (d *Data) GetPricesForMarket(name, quoteSymbol string) (Prices, error) {
	var price Prices

	coins, ok := d.raw[name]
	if !ok {
		return price, errors.New("market is not found")
	}

	prices, err := coins.GetPrices(quoteSymbol)
	if err != nil {
		return price, errors.Wrap(err)
	}

	return prices, nil
}

func (d *Data) GetPriceForMarket(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	prices, err := d.GetPricesForMarket(name, quoteSymbol)
	if err != nil {
		return decimal.Decimal{}, errors.Wrap(err)
	}

	return prices.GetPrice(baseSymbol)
}

type Markets map[string]Coins

type Coins map[string]Prices

type Prices []*domain.Price

func (c Coins) GetPriceForShotPath(baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	shotPath := graph.NewShopPath()

	prices := make([]*domain.Price, 0)
	for _, v := range c {
		var price []*domain.Price
		copy(price, v)
		prices = append(prices, price...)
	}

	fmt.Println(shotPath)

	return decimal.Decimal{}, nil
}

func (c Coins) GetPrices(quoteSymbol string) (Prices, error) {
	prices, ok := c[quoteSymbol]
	if !ok {
		return nil, errors.New("coins is not found")
	}

	return prices, nil
}

func (p Prices) GetPrice(baseSymbol string) (decimal.Decimal, error) {
	for _, v := range p {
		if v.BaseSymbol == baseSymbol {
			return v.Price, nil
		}
	}

	return decimal.Decimal{}, errors.New("price is not found")
}
