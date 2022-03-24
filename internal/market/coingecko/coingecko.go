package coingecko

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
	url            = "https://api.coingecko.com/api/v3"
	simplePriceUrl = url + "/simple/price"
	CoinsListUrl   = url + "/coins/list"
	BusinessName   = "coingecko"
)

type Coingecko struct {
	client       *resty.Client
	idToSymbol   map[string]string
	quoteSymbols []string
}

type id struct {
	ID     string
	Symbol string
}

func NewCoingecko(eConfig *config.ExchangeConfig) (*Coingecko, error) {
	baseSymbols, quoteSymbols := make([]string, len(eConfig.BaseSymbols)), make([]string, len(eConfig.QuoteSymbols))
	copy(baseSymbols, eConfig.BaseSymbols)
	copy(quoteSymbols, eConfig.QuoteSymbols)

	coingecko := &Coingecko{
		client: resty.New(),
	}
	err := coingecko.setSymbolToIds(config.StringLowerUpperForSlice(baseSymbols, config.StringLower))
	if err != nil {
		return nil, err
	}

	coingecko.setQuoteSymbol(config.StringLowerUpperForSlice(quoteSymbols, config.StringLower))

	return coingecko, nil
}

func (cg *Coingecko) GetName() string {
	return BusinessName
}

func (cg *Coingecko) GetPrices() (map[string][]*domain.Price, error) {
	resp, err := cg.client.R().
		SetQueryParams(map[string]string{
			"ids":           cg.getIdsToString(),
			"vs_currencies": cg.getQuoteSymbolsToString(),
		}).
		Get(simplePriceUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	raw := make(map[string]map[string]float64, 0)
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	prices := map[string][]*domain.Price{}
	for baseSymbol, quoteSymbols := range raw {
		for quoteSymbol, price := range quoteSymbols {
			quoteSymbolUpper := strings.ToUpper(quoteSymbol)
			if _, ok := prices[quoteSymbolUpper]; !ok {
				prices[quoteSymbolUpper] = make([]*domain.Price, 0)
			}

			prices[quoteSymbolUpper] = append(prices[quoteSymbolUpper], &domain.Price{
				BaseSymbol:  strings.ToUpper(cg.idToSymbol[baseSymbol]),
				QuoteSymbol: quoteSymbolUpper,
				Price:       decimal.NewFromFloat(price),
			})
		}
	}

	return prices, nil
}

func (cg *Coingecko) setQuoteSymbol(quoteSymbols []string) {
	cg.quoteSymbols = quoteSymbols
}

func (cg *Coingecko) setSymbolToIds(baseSymbols []string) error {
	if len(baseSymbols) == 0 {
		return errors.RecordNotFound
	}

	ids, err := cg.getIds()
	if err != nil {
		return errors.Wrap(err)
	}

	cg.idToSymbol = make(map[string]string, len(baseSymbols))
	for _, v := range baseSymbols {
		if i, ok := ids[v]; ok {
			cg.idToSymbol[i.ID] = v
		}
	}

	return nil
}

func (cg *Coingecko) getIds() (map[string]*id, error) {
	resp, err := cg.client.R().
		SetQueryParams(map[string]string{}).
		Get(CoinsListUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var ids []*id
	err = json.Unmarshal(resp.Body(), &ids)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	idsMap := make(map[string]*id, len(ids))
	for _, v := range ids {
		idsMap[v.Symbol] = v
	}

	return idsMap, nil
}

func (cg *Coingecko) getQuoteSymbolsToString() string {
	return strings.Join(cg.quoteSymbols, ",")
}

func (cg *Coingecko) getIdsToString() string {
	return strings.Join(cg.getIdsToSlice(), ",")
}

func (cg *Coingecko) getIdsToSlice() []string {
	ids := make([]string, 0, len(cg.idToSymbol))
	for k, _ := range cg.idToSymbol {
		ids = append(ids, k)
	}
	return ids
}
