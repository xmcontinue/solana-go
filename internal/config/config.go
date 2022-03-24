package config

import (
	"strings"

	aConfig "git.cplus.link/go/akit/config"
)

const (
	StringLower = int8(0)
	StringUpper = int8(1)
)

var (
	defaultBaseSymbols  = []string{"BTC", "ETH", "USDT", "BNB", "USDC", "XRP", "ADA", "LUNA", "SOL", "DOT", "AVAX", "DOGE", "BUSD", "UST", "SHIB", "WBTC", "CRO", "MATIC", "DAI", "LTC", "STETH", "ATOM", "NEAR", "LINK", "BCH", "TRX", "FTT", "ETC", "LEO", "ALGO", "OKB", "XLM", "AXS", "UNI", "HBAR", "ICP", "EGLD", "MANA", "SAND", "VET", "XMR", "FIL", "FTM", "CETH", "THETA", "KLAY", "XTZ", "WAVES", "OSMO", "FRAX", "MIM", "GRT", "CUSDC", "RUNE", "HNT", "EOS", "APE", "FLOW", "ZEC", "MIOTA", "AAVE", "CDAI", "CAKE", "GALA", "TFUEL", "MKR", "ONE", "BTT", "BSV", "HBTC", "QNT", "NEO", "AR", "OMI", "XRD", "XEC", "ENJ", "JUNO", "KSM", "STX", "HT", "KCS", "CEL", "DASH", "TUSD", "LRC", "HEART", "CELO", "AMP", "BAT", "NEXO", "CHZ", "CVX", "SNX", "MINA", "KDA", "BIT", "FXS", "XIDO", "GT"}
	defaultQuoteSymbols = []string{"USD"}
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
	defaultSlice := make([]string, len(defaultBaseSymbols))
	copy(defaultSlice, defaultBaseSymbols)
	e.BaseSymbols = SliceRemoveDuplicates(append(defaultSlice, StringLowerUpperForSlice(b, StringUpper)...))
}

func (e *ExchangeConfig) setQuoteSymbols(q []string) {
	defaultSlice := make([]string, len(defaultQuoteSymbols))
	copy(defaultSlice, defaultQuoteSymbols)
	e.QuoteSymbols = SliceRemoveDuplicates(append(defaultSlice, StringLowerUpperForSlice(q, StringUpper)...))
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

func SliceRemoveDuplicates(slice []string) []string {
	i := 0
	var j int
	for {
		if i >= len(slice)-1 {
			break
		}

		for j = i + 1; j < len(slice) && slice[i] == slice[j]; j++ {
		}
		slice = append(slice[:i+1], slice[j:]...)
		i++
	}
	return slice
}
