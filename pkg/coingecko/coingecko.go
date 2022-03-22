package coingecko

import (
	"encoding/json"
	"strings"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	url         = "https://api.coingecko.com/api/v3"
	tokenPrice  = "/simple/token_price/solana"
	simplePrice = "/simple/price"
)

var (
	client         = resty.New()
	prices         = make(map[string]decimal.Decimal)
	priceForSymbol = make(map[string]decimal.Decimal)
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
	newPriceForSymbol := make(map[string]decimal.Decimal, 0)

	f := func(tokenConf domain.Token) {
		if _, ok := newPrices[tokenConf.SwapTokenAccount]; !ok {
			var (
				err   error
				price = decimal.NewFromInt(1)
			)

			if strings.Contains(strings.ToUpper(tokenConf.Symbol), "SOL") {
				price, err = GetPriceFromIds("solana")
			} else {
				price, err = GetPriceFromTokenMintAccount(tokenConf.TokenMint)
			}

			if err != nil {
				newPrices[tokenConf.SwapTokenAccount] = decimal.NewFromInt(1)
			}
			newPrices[tokenConf.SwapTokenAccount] = price

			if _, ok = newPriceForSymbol[tokenConf.Symbol]; !ok {
				newPriceForSymbol[tokenConf.Symbol] = price
			}
		}
	}

	for _, v := range keys {
		f(v.TokenA)
		f(v.TokenB)
	}

	prices = newPrices
	priceForSymbol = newPriceForSymbol
}

// GetPriceFromTokenMintAccount 通过token mint账户地址拿取币价
func GetPriceFromTokenMintAccount(tokenMintAccount string) (decimal.Decimal, error) {
	defaultPrice := decimal.NewFromInt(1)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"contract_addresses": tokenMintAccount,
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

	priceStruct, ok := resMap[tokenMintAccount]
	if !ok {
		return defaultPrice, nil
	}

	return priceStruct.Usd, nil
}

// GetPriceFromIds 通过ids拿取币价
func GetPriceFromIds(ids string) (decimal.Decimal, error) {
	defaultPrice := decimal.NewFromInt(1)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"ids":           ids,
			"vs_currencies": "usd",
		}).
		Get(url + simplePrice)
	if err != nil {
		return defaultPrice, errors.Wrap(err)
	}

	var resMap map[string]Price

	err = json.Unmarshal(resp.Body(), &resMap)
	if err != nil {
		return defaultPrice, errors.Wrap(err)
	}

	priceStruct, ok := resMap[ids]
	if !ok {
		return defaultPrice, errors.New("get price error")
	}

	return priceStruct.Usd, nil
}

// GetPriceForTokenAccount ...
func GetPriceForTokenAccount(tokenAccount string) (price decimal.Decimal) {
	price, ok := prices[tokenAccount]
	if !ok {
		price = decimal.NewFromInt(1)
	}
	return price
}

// GetPriceForSymbol ...
func GetPriceForSymbol(symbol string) (price decimal.Decimal) {
	price, ok := priceForSymbol[symbol]
	if !ok {
		price = decimal.NewFromInt(1)
	}
	return price
}
