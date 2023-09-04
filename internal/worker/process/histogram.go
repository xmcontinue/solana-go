package process

import (
	"encoding/json"
	"reflect"
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type Price struct {
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Settle decimal.Decimal
	Avg    decimal.Decimal
	Date   *time.Time
}

func (s *Price) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Price) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type SwapHistogram struct {
	Tvl  decimal.Decimal
	Vol  decimal.Decimal
	Date *time.Time
}

type SwapHistogramNumber struct {
	Num  decimal.Decimal `json:"num"`
	Date *time.Time      `json:"date"`
}

func (s *SwapHistogram) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SwapHistogram) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *SwapHistogram) IsEmpty() bool {
	return reflect.DeepEqual(s, SwapHistogram{})
}

type HistogramZ struct {
	Score  int64
	Member *SwapHistogram
}

type PriceZ struct {
	Score  int64
	Member *Price
}
