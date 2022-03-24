package config

import (
	"strings"

	aConfig "git.cplus.link/go/akit/config"
)

const (
	StringLower = int8(0)
	StringUpper = int8(1)
)

type ExchangeConfig struct {
	BaseSymbols    []string
	QuoteSymbols   []string
	ReplaceSymbols map[string]string
}

func NewExchangeConfigForViper(viperConf *aConfig.Config) (*ExchangeConfig, error) {
	var (
		baseSymbols    string
		quoteSymbols   string
		replaceSymbols map[string]string
		exchangeConfig = &ExchangeConfig{}
	)
	err := viperConf.UnmarshalKey("exchange.base_symbols", &baseSymbols)
	if err != nil {
		return nil, err
	}
	exchangeConfig.setBaseSymbols(strings.Split(baseSymbols, ","))

	err = viperConf.UnmarshalKey("exchange.quote_symbols", &quoteSymbols)
	if err != nil {
		return nil, err
	}
	exchangeConfig.setQuoteSymbols(strings.Split(quoteSymbols, ","))

	err = viperConf.UnmarshalKey("exchange.replace_symbols", &replaceSymbols)
	if err != nil {
		return nil, err
	}
	exchangeConfig.setReplaceSymbols(replaceSymbols)

	return exchangeConfig, nil
}

func (e *ExchangeConfig) setBaseSymbols(b []string) {
	e.BaseSymbols = StringLowerUpperForSlice(b, StringUpper)
}

func (e *ExchangeConfig) setQuoteSymbols(q []string) {
	e.QuoteSymbols = StringLowerUpperForSlice(q, StringUpper)
}

func (e *ExchangeConfig) setReplaceSymbols(r map[string]string) {
	e.ReplaceSymbols = StringLowerUpperForMap(r, StringUpper)
}

func StringLowerUpperForSlice(l []string, typ int8) []string {
	for k, v := range l {
		l[k] = StringLowerUpper(v, typ)
	}
	return l
}

func StringLowerUpper(s string, typ int8) string {
	if typ == 0 {
		return strings.ToLower(s)
	}
	return strings.ToUpper(s)
}

func StringLowerUpperForMap(l map[string]string, typ int8) map[string]string {
	res := make(map[string]string, len(l))
	for k, v := range l {
		res[StringLowerUpper(k, typ)] = StringLowerUpper(v, typ)
	}
	return res
}
