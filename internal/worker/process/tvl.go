package process

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type SymbolPri struct {
	Date time.Time
	Num  decimal.Decimal
}

func (s *SymbolPri) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SymbolPri) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *SymbolPri) IsEmpty() bool {
	return reflect.DeepEqual(s, SymbolPri{})
}

func tvlOfToken() error {
	var now = time.Now()

	symBolMap := make(map[string]bool, 10)
	for _, swapConfig := range sol.SwapConfigList() {
		symBolMap[swapConfig.TokenA.Symbol] = true
		symBolMap[swapConfig.TokenB.Symbol] = true
	}

	ctx := context.Background()

	for symbol := range symBolMap {
		filters := []model.Filter{
			model.NewFilter("token_a_symbol = ? or token_b_symbol = ?", symbol, symbol),
			model.NewFilter("date_type = ?", domain.DateHour),
			model.OrderFilter("date desc"),
		}

		tokenTvlPriceInUSDs, err := model.SumTvlPriceInUSD(ctx, 24, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(tokenTvlPriceInUSDs) == 0 {
			continue
		}

		tokenKey := domain.TokenKey(symbol)

		// 减少for 循环
		dateAndPriMap := make(map[int64]*model.DateAndPrice, len(tokenTvlPriceInUSDs))
		for index := range tokenTvlPriceInUSDs {
			dateAndPriMap[tokenTvlPriceInUSDs[index].Date.Unix()] = tokenTvlPriceInUSDs[index]
		}

		now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, tokenTvlPriceInUSDs[0].Date.Location())
		begin := now.Add(-23 * time.Hour)
		avgList := make([]*SymbolPri, 24, 24)
		for index := range avgList {
			avgList[index] = &SymbolPri{
				Date: begin.Add(time.Hour * time.Duration(index)),
			}
		}

		// 找到第一个数据
		lastAvg := &model.DateAndPrice{}
		for index := range tokenTvlPriceInUSDs {
			if tokenTvlPriceInUSDs[len(tokenTvlPriceInUSDs)-index-1].Date.After((avgList)[0].Date) {
				break
			}
			lastAvg = tokenTvlPriceInUSDs[len(tokenTvlPriceInUSDs)-index-1]
		}

		for index, avg := range avgList {
			lastSwapCountKLine, ok := dateAndPriMap[avg.Date.Unix()]
			if ok {
				lastAvg = lastSwapCountKLine
				avgList[index].Num = lastSwapCountKLine.Tvl
			} else {
				avgList[index].Num = lastAvg.Tvl // 上一个周期的结束值用作空缺周期的平均值
			}
		}

		// 去掉列表前面的零值
		for i, v := range avgList {
			if !v.Num.IsZero() {
				avgList = avgList[i:]
				break
			}
		}

		// lua 通过脚本更新
		newZ := make([]interface{}, 0, len(avgList)+1)
		for i := range avgList {
			newZ = append(newZ, avgList[i].Date.Unix())
			newZ = append(newZ, avgList[i])
		}

		_, err = delAndAddByZSet.Run(ctx, redisClient, []string{tokenKey}, newZ).Result()
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}
