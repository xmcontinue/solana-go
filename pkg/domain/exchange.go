package domain

import (
	"git.cplus.link/go/akit/util/decimal"
)

var ApiHost string

type Price struct {
	BaseSymbol  string          `json:"base_symbol"`
	QuoteSymbol string          `json:"quote_symbol"`
	Price       decimal.Decimal `json:"price"`
}

func SetApiHost(host string) {
	if host != "" {
		ApiHost = host
	}
}