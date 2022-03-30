package ftx

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	url            = "https://ftx.us/api"
	simplePriceUrl = url + "/markets"
	BusinessName   = "ftx"
)

type Ftx struct {
	client       *resty.Client
	baseSymbols  map[string]struct{}
	quoteSymbols map[string]struct{}
}

func NewFtx(eConfig *config.ExchangeConfig) *Ftx {
	baseSymbols, quoteSymbols := eConfig.GetBaseSymbolsForCopy(), eConfig.GetQuoteSymbolsForCopy()

	ftx := &Ftx{
		client: resty.New(),
	}

	ftx.setQuoteSymbol(quoteSymbols)

	ftx.setBaseSymbol(baseSymbols)

	return ftx
}

func (f *Ftx) GetName() string {
	return BusinessName
}

func (f *Ftx) GetPrices() (map[string][]*domain.Price, error) {
	resp, err := f.client.R().
		SetQueryParams(map[string]string{}).
		Get(simplePriceUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var raw struct {
		Result []struct {
			BaseCurrency  string          `json:"baseCurrency"`
			QuoteCurrency string          `json:"quoteCurrency"`
			Last          decimal.Decimal `json:"last"`
		}
	}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	prices := map[string][]*domain.Price{}
	for k, _ := range f.quoteSymbols {
		prices[k] = make([]*domain.Price, 0)
	}
	for _, v := range raw.Result {
		if v.QuoteCurrency == "USDT" {
			v.QuoteCurrency = "USD"
		}
		if f.hasQuoteSymbol(v.QuoteCurrency) && f.hasBaseSymbol(v.BaseCurrency) {
			prices[v.QuoteCurrency] = append(prices[v.QuoteCurrency], &domain.Price{
				BaseSymbol:  v.BaseCurrency,
				QuoteSymbol: v.QuoteCurrency,
				Price:       v.Last,
			})
		}
	}

	return prices, nil
}

func (f *Ftx) setQuoteSymbol(quoteSymbols []string) {
	quoteSymbolsM := make(map[string]struct{}, 0)
	for _, v := range quoteSymbols {
		quoteSymbolsM[v] = struct{}{}
	}

	f.quoteSymbols = quoteSymbolsM
}

func (f *Ftx) setBaseSymbol(baseSymbols []string) {
	baseSymbolsM := make(map[string]struct{}, 0)
	for _, v := range baseSymbols {
		baseSymbolsM[v] = struct{}{}
	}

	f.baseSymbols = baseSymbolsM
}
func (f *Ftx) hasQuoteSymbol(quoteSymbol string) bool {
	_, ok := f.quoteSymbols[quoteSymbol]
	return ok
}

func (f *Ftx) hasBaseSymbol(baseSymbol string) bool {
	_, ok := f.baseSymbols[baseSymbol]
	return ok
}
