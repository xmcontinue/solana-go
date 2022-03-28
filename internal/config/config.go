package config

import (
	"encoding/json"
	"strings"

	aConfig "git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
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
	BaseSymbols    []string          `json:"base_symbols"            yaml:"base_symbols"            mapstructure:"base_symbols"`
	QuoteSymbols   []string          `json:"quote_symbols"           yaml:"quote_symbols"           mapstructure:"quote_symbols"`
	ReplaceSymbols map[string]string `json:"replace_symbols"         yaml:"replace_symbols"         mapstructure:"replace_symbols"`
}

func NewExchangeConfig() *ExchangeConfig {
	return &ExchangeConfig{}
}

func NewExchangeConfigForViper(viperConf *aConfig.Config) (*ExchangeConfig, error) {
	exchangeConfig := NewExchangeConfig()

	err := viperConf.UnmarshalKey("exchange", &exchangeConfig)
	if err != nil {
		return exchangeConfig, errors.Wrap(err)
	}

	exchangeConfig.LoadConfig()

	return exchangeConfig, nil
}

func (e *ExchangeConfig) LoadConfig() {
	e.loadBaseSymbols()
	e.loadQuoteSymbols()
	e.loadReplaceSymbols()
}

func (e *ExchangeConfig) loadBaseSymbols() {
	defaultSlice := make([]string, len(defaultBaseSymbols))
	copy(defaultSlice, defaultBaseSymbols)
	e.BaseSymbols = SliceRemoveDuplicates(append(append(defaultSlice, e.QuoteSymbols...), StringLowerUpperForSlice(e.BaseSymbols, StringUpper)...))
}

func (e *ExchangeConfig) loadQuoteSymbols() {
	defaultSlice := make([]string, len(defaultQuoteSymbols))
	copy(defaultSlice, defaultQuoteSymbols)
	e.QuoteSymbols = SliceRemoveDuplicates(append(defaultSlice, StringLowerUpperForSlice(e.QuoteSymbols, StringUpper)...))
}

func (e *ExchangeConfig) loadReplaceSymbols() {
	e.ReplaceSymbols = StringLowerUpperForMap(e.ReplaceSymbols, StringUpper)
}

func (e *ExchangeConfig) setBaseSymbols(b []string) {
	defaultSlice := make([]string, len(defaultBaseSymbols))
	copy(defaultSlice, defaultBaseSymbols)
	e.BaseSymbols = SliceRemoveDuplicates(append(append(defaultSlice, e.QuoteSymbols...), StringLowerUpperForSlice(b, StringUpper)...))
}

func (e *ExchangeConfig) setQuoteSymbols(q []string) {
	defaultSlice := make([]string, len(defaultQuoteSymbols))
	copy(defaultSlice, defaultQuoteSymbols)
	e.QuoteSymbols = SliceRemoveDuplicates(append(defaultSlice, StringLowerUpperForSlice(q, StringUpper)...))
}

func (e *ExchangeConfig) setReplaceSymbols(r map[string]string) {
	e.ReplaceSymbols = StringLowerUpperForMap(r, StringUpper)
}

func (e *ExchangeConfig) Equal(config *ExchangeConfig) bool {
	c1Json, _ := json.Marshal(e)
	c2Json, _ := json.Marshal(config)
	return string(c1Json) == string(c2Json)
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
