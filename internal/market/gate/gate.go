package gate

import (
	"encoding/json"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	url            = "https://api.gateio.ws/api/v4/spot"
	simplePriceUrl = url + "/tickers"
	CoinsListUrl   = url + "/v2/public/symbols"
	BusinessName   = "gate"
)

type Gate struct {
	client       *resty.Client
	baseSymbols  map[string]struct{}
	quoteSymbols map[string]struct{}
}

func NewGate(eConfig *config.ExchangeConfig) *Gate {
	baseSymbols, quoteSymbols := eConfig.GetBaseSymbolsForCopy(), eConfig.GetQuoteSymbolsForCopy()

	gate := &Gate{
		client: resty.New(),
	}

	gate.setQuoteSymbol(quoteSymbols)

	gate.setBaseSymbol(baseSymbols)

	return gate
}

func (g *Gate) GetName() string {
	return BusinessName
}

func (g *Gate) GetPrices() (map[string][]*domain.Price, error) {
	resp, err := g.client.R().
		SetQueryParams(map[string]string{}).
		Get(simplePriceUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var raw []struct {
		CurrencyPair string          `json:"currency_pair"`
		Last         decimal.Decimal `json:"last"`
	}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	prices := map[string][]*domain.Price{}
	for k, _ := range g.quoteSymbols {
		prices[k] = make([]*domain.Price, 0)
	}
	for _, v := range raw {
		symbols := strings.Split(v.CurrencyPair, "_")
		if len(symbols) != 2 {
			continue
		}
		baseSymbol, quoteSymbol := symbols[0], symbols[1]
		if quoteSymbol == "USDT" {
			quoteSymbol = "USD"
		}
		if g.hasQuoteSymbol(quoteSymbol) && g.hasBaseSymbol(baseSymbol) {
			prices[quoteSymbol] = append(prices[quoteSymbol], &domain.Price{
				BaseSymbol:  baseSymbol,
				QuoteSymbol: quoteSymbol,
				Price:       v.Last,
			})
		}
	}

	return prices, nil
}

func (g *Gate) setQuoteSymbol(quoteSymbols []string) {
	quoteSymbolsM := make(map[string]struct{}, 0)
	for _, v := range quoteSymbols {
		quoteSymbolsM[v] = struct{}{}
	}

	g.quoteSymbols = quoteSymbolsM
}

func (g *Gate) setBaseSymbol(baseSymbols []string) {
	baseSymbolsM := make(map[string]struct{}, 0)
	for _, v := range baseSymbols {
		baseSymbolsM[v] = struct{}{}
	}

	g.baseSymbols = baseSymbolsM
}
func (g *Gate) hasQuoteSymbol(quoteSymbol string) bool {
	_, ok := g.quoteSymbols[quoteSymbol]
	return ok
}

func (g *Gate) hasBaseSymbol(baseSymbol string) bool {
	_, ok := g.baseSymbols[baseSymbol]
	return ok
}
