package crema

import (
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	tokenPrice = "/price"
)

var (
	client = resty.New()
)

type Res struct {
	Data struct {
		Prices []Price `json:"prices"`
	} `json:"data"`
}

type Price struct {
	BaseSymbol  string          `json:"base_symbol"`
	QuoteSymbol string          `json:"quote_symbol"`
	Price       decimal.Decimal `json:"price"`
}

// GetPriceForSymbol ...
func GetPriceForSymbol(symbol string) (price decimal.Decimal, err error) {
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"base_symbol":  symbol,
			"quote_symbol": "usd",
			"market":       "avg",
		}).
		Get(domain.ApiHost + tokenPrice)
	if err != nil {
		return decimal.NewFromInt(1), errors.Wrap(err)
	}

	var resMap Res
	err = json.Unmarshal(resp.Body(), &resMap)
	if err != nil {
		return decimal.NewFromInt(1), errors.Wrap(err)
	}

	if len(resMap.Data.Prices) == 0 {
		return decimal.NewFromInt(1), errors.RecordNotFound
	}

	return resMap.Data.Prices[0].Price, nil
}

// GetPriceForBaseSymbol ...
func GetPriceForBaseSymbol(symbol string) (price decimal.Decimal, err error) {
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"base_symbol":  symbol,
			"quote_symbol": "usd",
		}).
		Get(domain.ApiHost + tokenPrice)
	if err != nil {
		return decimal.NewFromInt(1), errors.Wrap(err)
	}

	var resMap Res
	err = json.Unmarshal(resp.Body(), &resMap)
	if err != nil {
		return decimal.NewFromInt(1), errors.Wrap(err)
	}

	if len(resMap.Data.Prices) == 0 {
		return decimal.NewFromInt(1), errors.RecordNotFound
	}

	return resMap.Data.Prices[0].Price, nil
}
