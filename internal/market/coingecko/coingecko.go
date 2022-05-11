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

type CoinGecko struct {
	client       *resty.Client
	idToSymbol   map[string]string
	quoteSymbols []string
}

type id struct {
	ID     string
	Symbol string
}

func NewCoinGecko(eConfig *config.ExchangeConfig) (*CoinGecko, error) {
	baseSymbols, quoteSymbols := eConfig.GetBaseSymbolsForCopy(), eConfig.GetQuoteSymbolsForCopy()

	coinGecko := &CoinGecko{
		client: resty.New(),
	}
	err := coinGecko.setSymbolToIds(config.StringLowerUpperForSlice(baseSymbols, config.StringLower))
	if err != nil {
		return nil, err
	}

	coinGecko.setQuoteSymbol(config.StringLowerUpperForSlice(quoteSymbols, config.StringLower))

	return coinGecko, nil
}

func (cg *CoinGecko) GetName() string {
	return BusinessName
}

func (cg *CoinGecko) GetPrices() (map[string][]*domain.Price, error) {
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

			if price != 0 {
				if strings.ToUpper(cg.idToSymbol[baseSymbol]) == quoteSymbolUpper {
					price = 1
				}

				prices[quoteSymbolUpper] = append(prices[quoteSymbolUpper], &domain.Price{
					BaseSymbol:  strings.ToUpper(cg.idToSymbol[baseSymbol]),
					QuoteSymbol: quoteSymbolUpper,
					Price:       decimal.NewFromFloat(price),
				})
			}
		}
	}

	return prices, nil
}

func (cg *CoinGecko) setQuoteSymbol(quoteSymbols []string) {
	cg.quoteSymbols = quoteSymbols
}

func (cg *CoinGecko) setSymbolToIds(baseSymbols []string) error {
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

func (cg *CoinGecko) getIds() (map[string]*id, error) {
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
		if v.ID == "project-pai" {
			continue
		}
		if v.ID == "uniswap-state-dollar" {
			continue
		}
		if v.ID == "usd-coin-avalanche-bridged-usdc-e" {
			continue
		}

		idsMap[v.Symbol] = v
	}
	idsMap["usdc"] = &id{
		"usd-coin",
		"usdc",
	}

	return idsMap, nil
}

func (cg *CoinGecko) getQuoteSymbolsToString() string {
	return strings.Join(cg.quoteSymbols, ",")
}

func (cg *CoinGecko) getIdsToString() string {
	return strings.Join(cg.getIdsToSlice(), ",")
}

func (cg *CoinGecko) getIdsToSlice() []string {
	ids := make([]string, 0, len(cg.idToSymbol))
	for k, _ := range cg.idToSymbol {
		ids = append(ids, k)
	}
	return ids
}
