package bybit

import (
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	url            = "https://api.bybit.com/"
	simplePriceUrl = url + "/v2/public/tickers"
	CoinsListUrl   = url + "/v2/public/symbols"
	BusinessName   = "bybit"
)

type ByBit struct {
	client       *resty.Client
	pairToSymbol map[string]*pair
	quoteSymbols map[string]struct{}
}

type pair struct {
	BaseSymbol  string
	QuoteSymbol string
}

func NewByBit(eConfig *config.ExchangeConfig) (*ByBit, error) {
	baseSymbols, quoteSymbols := eConfig.GetBaseSymbolsForCopy(), eConfig.GetQuoteSymbolsForCopy()

	coinGecko := &ByBit{
		client: resty.New(),
	}

	coinGecko.setQuoteSymbol(quoteSymbols)

	err := coinGecko.setPairToSymbol(baseSymbols)
	if err != nil {
		return nil, err
	}

	return coinGecko, nil
}

func (b *ByBit) GetName() string {
	return BusinessName
}

func (b *ByBit) GetPrices() (map[string][]*domain.Price, error) {
	resp, err := b.client.R().
		SetQueryParams(map[string]string{}).
		Get(simplePriceUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var raw struct {
		Result []struct {
			Symbol    string          `json:"symbol"`
			LastPrice decimal.Decimal `json:"last_price"`
		}
	}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	prices := map[string][]*domain.Price{}
	for k, _ := range b.quoteSymbols {
		prices[k] = make([]*domain.Price, 0)
	}
	for _, v := range raw.Result {
		p, ok := b.pairToSymbol[v.Symbol]
		if !ok {
			continue
		}

		if p.BaseSymbol == p.QuoteSymbol {
			v.LastPrice = decimal.NewFromInt(1)
		}

		if _, ok = prices[p.QuoteSymbol]; ok {
			prices[p.QuoteSymbol] = append(prices[p.QuoteSymbol], &domain.Price{
				BaseSymbol:  p.BaseSymbol,
				QuoteSymbol: p.QuoteSymbol,
				Price:       v.LastPrice,
			})
		}
	}

	return prices, nil
}

func (b *ByBit) setQuoteSymbol(quoteSymbols []string) {
	quoteSymbolsM := make(map[string]struct{}, 0)
	for _, v := range quoteSymbols {
		quoteSymbolsM[v] = struct{}{}
	}

	b.quoteSymbols = quoteSymbolsM
}

func (b *ByBit) setPairToSymbol(baseSymbols []string) error {
	if len(baseSymbols) == 0 {
		return errors.RecordNotFound
	}

	baseSymbolsM := make(map[string]struct{}, 0)
	for _, v := range baseSymbols {
		baseSymbolsM[v] = struct{}{}
	}

	var err error
	b.pairToSymbol, err = b.getPairs(baseSymbolsM)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (b *ByBit) getPairs(baseSymbolsM map[string]struct{}) (map[string]*pair, error) {
	resp, err := b.client.R().
		SetQueryParams(map[string]string{}).
		Get(CoinsListUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var res struct {
		Result []struct {
			Name          string `json:"name"`
			BaseCurrency  string `json:"base_currency"`
			QuoteCurrency string `json:"quote_currency"`
		}
	}
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	pairs := make(map[string]*pair, len(res.Result))
	for _, v := range res.Result {
		quoteSymbol := v.QuoteCurrency
		if quoteSymbol == "USDT" {
			quoteSymbol = "USD"
		}

		if _, ok := b.quoteSymbols[quoteSymbol]; !ok {
			continue
		}

		if _, ok := baseSymbolsM[v.BaseCurrency]; !ok {
			continue
		}
		pairs[v.Name] = &pair{
			v.BaseCurrency,
			quoteSymbol,
		}
	}

	return pairs, nil
}
