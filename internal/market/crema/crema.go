package crema

import (
	"encoding/json"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	simplePriceUrl = "/v1/swap/count"
	BusinessName   = "crema"
)

type Crema struct {
	client *resty.Client
}

func NewCrema() *Crema {
	return &Crema{
		client: resty.New(),
	}
}

func (c *Crema) GetName() string {
	return BusinessName
}

type Raw struct {
	Data struct {
		Tokens []*Token `json:"tokens"`
	} `json:"data"`
}
type Token struct {
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"`
}

func (c *Crema) GetPrices() (map[string][]*domain.Price, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{}).
		Get(domain.ApiHost + simplePriceUrl)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var raw Raw
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	prices := map[string][]*domain.Price{}
	for _, token := range raw.Data.Tokens {
		prices["USD"] = append(prices["USD"], &domain.Price{
			BaseSymbol:  strings.ToUpper(token.Name),
			QuoteSymbol: "USD",
			Price:       token.Price,
		})
	}

	return prices, nil
}
