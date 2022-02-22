package coingecko

import (
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"
)

const (
	url        = "https://api.coingecko.com/api/v3"
	tokenPrice = "/simple/token_price/solana"
)

var (
	client = resty.New()
)

type Price struct {
	Usd decimal.Decimal `json:"usd"`
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
