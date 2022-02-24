package coingecko

import (
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/chain/sol"
)

const (
	url        = "https://api.coingecko.com/api/v3"
	tokenPrice = "/simple/token_price/solana"
)

var (
	client = resty.New()
	prices = make(map[string]decimal.Decimal)
)

type Price struct {
	Usd decimal.Decimal `json:"usd"`
}

func Init() {
	syncPrice()
	go func() {
		for {
			time.Sleep(time.Minute)
			syncPrice()
		}
	}()
}

func syncPrice() {
	keys := sol.SwapConfigList()
	newPrices := make(map[string]decimal.Decimal, 0)
	for _, v := range keys {

		if _, ok := newPrices[v.TokenA.SwapTokenAccount]; !ok {
			tokenAPrice, err := GetPriceFromTokenAccount(v.TokenA.SwapTokenAccount)
			if err != nil {
				newPrices[v.TokenA.SwapTokenAccount] = decimal.NewFromInt(1)
			}
			newPrices[v.TokenA.SwapTokenAccount] = tokenAPrice
		}

		if _, ok := newPrices[v.TokenB.SwapTokenAccount]; !ok {
			tokenAPrice, err := GetPriceFromTokenAccount(v.TokenB.SwapTokenAccount)
			if err != nil {
				newPrices[v.TokenB.SwapTokenAccount] = decimal.NewFromInt(1)
			}
			newPrices[v.TokenB.SwapTokenAccount] = tokenAPrice
		}

	}

	prices = newPrices
}

// GetPriceFromTokenAccount 通过token账户地址拿取币价
func GetPriceFromTokenAccount(tokenAccount string) (decimal.Decimal, error) {
	defaultPrice := decimal.NewFromInt(1)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"contract_addresses": tokenAccount,
			"vs_currencies":      "usd",
		}).
		Get(url + tokenPrice)
	if err != nil {
		return defaultPrice, nil
	}

	var resMap map[string]Price

	err = json.Unmarshal(resp.Body(), &resMap)
	if err != nil {
		return defaultPrice, errors.Wrap(err)
	}

	priceStruct, ok := resMap[tokenAccount]
	if !ok {
		return defaultPrice, nil
	}

	return priceStruct.Usd, nil
}

// GetPriceForCache ...
func GetPriceForCache(tokenAccount string) (price decimal.Decimal) {
	price, ok := prices[tokenAccount]
	if !ok {
		price = decimal.NewFromInt(1)
	}
	return price
}
