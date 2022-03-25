package exchanger

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/exchanger/graph"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/errcode"
)

type Data struct {
	raw *Markets
	avg *Coins
}

func NewData() *Data {
	return &Data{}
}

func (d *Data) LoadRawData(raw map[string]map[string][]*domain.Price) {
	r := make(Markets, len(raw))
	for p, qs := range raw {
		r[p] = NewCoins()
		r[p].LoadRawData(qs)
	}
	d.raw = &r
}

func (d *Data) GetPriceForMarketForShotPath(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	var price decimal.Decimal

	coins, ok := (*d.raw)[name]
	if !ok {
		return price, errors.New("market is not found")
	}

	price, err := coins.GetPriceForShotPath(baseSymbol, quoteSymbol)
	if err != nil {
		return price, errors.Wrap(err)
	}

	return price, nil
}

func (d *Data) GetPricesForMarket(name, quoteSymbol string) (*Prices, error) {
	var price *Prices

	coins, ok := (*d.raw)[name]
	if !ok {
		return price, errors.New("market is not found")
	}

	prices, err := coins.GetPrices(quoteSymbol)
	if err != nil {
		return price, errors.Wrap(err)
	}

	return prices, nil
}

func (d *Data) GetPriceForMarket(name, baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	prices, err := d.GetPricesForMarket(name, quoteSymbol)
	if err != nil {
		return decimal.Decimal{}, errors.Wrap(err)
	}

	return prices.GetPrice(baseSymbol)
}

type Markets map[string]*Coins

type Coins struct {
	prices     map[string]*Prices
	graph      graph.Graph
	currencies *Currencies
}

func NewCoins() *Coins {
	return &Coins{}
}

func (c *Coins) LoadRawData(raw map[string][]*domain.Price) {
	c.loadPrices(raw)
	c.loadGraph()
	c.loadCurrencies()
}

func (c *Coins) loadPrices(raw map[string][]*domain.Price) {
	coins := make(map[string]*Prices, len(raw))
	for k, v := range raw {
		coins[k] = NewPrices()
		coins[k].LoadRawData(v)
	}
	c.prices = coins
}

func (c *Coins) loadGraph() {
	c.graph = graph.NewGraphFromMap(c.ToMap(), true)
}

func (c *Coins) loadCurrencies() {
	c.currencies = NewCurrencies()
	c.currencies.LoadRawData(c.ToSlice())
}

func (c *Coins) ToMap() *map[string]*[]*domain.Price {
	res := make(map[string]*[]*domain.Price, len(c.GetAllPrices()))
	for k, v := range c.GetAllPrices() {
		res[k] = (*[]*domain.Price)(v)
	}
	return &res
}

func (c *Coins) ToSlice() *[]*domain.Price {
	res := make([]*domain.Price, 0, len(c.GetAllPrices()))
	for _, price := range c.GetAllPrices() {
		for _, v := range *price {
			res = append(res, v)
		}
	}
	return &res
}

func (c Coins) GetPriceForShotPath(baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	if !c.GetCurrencies().Has(baseSymbol) || !c.GetCurrencies().Has(quoteSymbol) {
		return decimal.Decimal{}, errors.RecordNotFound
	}

	ids, _, err := graph.Dijkstra(c.GetGraph(), graph.NewStringID(baseSymbol), graph.NewStringID(quoteSymbol))

	if err != nil {
		return decimal.Decimal{}, errcode.GetPriceFailed
	}

	return c.GetPriceForIDs(ids), nil
}

func (c Coins) GetPriceForIDs(ids []graph.ID) decimal.Decimal {
	price := decimal.NewFromInt(1)

	l := len(ids) - 1
	for i := 0; i < l; i++ {
		p, _ := c.GetPriceForPair(ids[i].String(), ids[i+1].String())
		price = price.Mul(p)
	}

	return price
}

func (c Coins) GetPriceForPair(baseSymbol, quoteSymbol string) (decimal.Decimal, error) {
	_, orderErr := c.GetPrices(quoteSymbol)

	if orderErr != nil {
		baseSymbol, quoteSymbol = quoteSymbol, baseSymbol
	}

	prices, err := c.GetPrices(quoteSymbol)
	if err != nil {
		return decimal.Decimal{}, errors.RecordNotFound
	}

	price, err := prices.GetPrice(baseSymbol)
	if err != nil {
		return decimal.Decimal{}, errors.RecordNotFound
	}

	if orderErr != nil {
		return decimal.NewFromInt(1).Div(price), nil
	}

	return price, nil
}

func (c Coins) GetPrices(quoteSymbol string) (*Prices, error) {
	prices, ok := c.GetAllPrices()[quoteSymbol]
	if !ok {
		return nil, errors.New("coins is not found")
	}

	return prices, nil
}

func (c Coins) GetAllPrices() map[string]*Prices {
	return c.prices
}

func (c Coins) GetGraph() graph.Graph {
	return c.graph
}

func (c Coins) GetCurrencies() *Currencies {
	return c.currencies
}

type Prices []*domain.Price

func NewPrices() *Prices {
	return &Prices{}
}

func (p *Prices) LoadRawData(raw []*domain.Price) {
	*p = raw
}

func (p Prices) GetPrice(baseSymbol string) (decimal.Decimal, error) {
	for _, v := range p {
		if v.BaseSymbol == baseSymbol {
			return v.Price, nil
		}
	}

	return decimal.Decimal{}, errors.New("price is not found")
}

type Currencies map[string]struct{}

func NewCurrencies() *Currencies {
	return &Currencies{}
}

func (c *Currencies) LoadRawData(raw *[]*domain.Price) {
	l := make(map[string]struct{}, len(*raw))
	for _, v := range *raw {
		l[v.BaseSymbol] = struct{}{}
		l[v.QuoteSymbol] = struct{}{}
	}
	*c = l
}

func (c *Currencies) Has(key string) bool {
	_, ok := (map[string]struct{})(*c)[key]
	return ok
}
