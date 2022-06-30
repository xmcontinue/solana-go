package crema

import (
	"encoding/json"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/internal/market/coingecko"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	simplePriceUrl = "/v1/swap/count"
	BusinessName   = "crema"
)

var newPair = map[string]string{
	"LDO":  "HZRCwxP2Vq9PCpPXooayhJ2bxTpo5xfpQrwB1svh332p",
	"MNDE": "MNDEFzGvMt87ueuHvVU9VcTqsAP5b3fTGPsHuuPA5ey",
}

type Crema struct {
	client    *resty.Client
	coingecko *coingecko.CoinGecko
}

func NewCrema(eConfig *config.ExchangeConfig) (*Crema, error) {
	coin, err := coingecko.NewCoinGecko(eConfig)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &Crema{
		client:    resty.New(),
		coingecko: coin,
	}, nil
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

	for k, v := range newPair {
		p, err := c.coingecko.GetPriceForContract(v)
		if err != nil {
			continue
		}
		prices["USD"] = append(prices["USD"], &domain.Price{
			BaseSymbol:  strings.ToUpper(k),
			QuoteSymbol: "USD",
			Price:       p,
		})
	}

	// 特殊代币
	// "HDG":  "5PmpMzWjraf3kSsGEKtqdUsCoLhptg4yriZ17LKKdBBy",
	hdgResp, err := c.client.R().
		SetQueryParams(map[string]string{}).
		Get("https://price.jup.ag/v1/price?id=HDG")
	var hdg struct {
		Data struct {
			Price float64 `json:"price"`
		} `json:"data"`
	}
	err = json.Unmarshal(hdgResp.Body(), &hdg)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	prices["USD"] = append(prices["USD"], &domain.Price{
		BaseSymbol:  "HDG",
		QuoteSymbol: "USD",
		Price:       decimal.NewFromFloat(hdg.Data.Price),
	})

	return prices, nil

}
