package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

// SwapCount 同步更新swap_counts表和user_swap_counts表
type SwapCount struct {
	ID                int64
	LastTransactionID int64
	SwapAccount       string
	SwapRecords       []*parse.SwapRecord
	tx                *parse.Tx
	BlockDate         *time.Time
	spec              string
}

// ParserDate 按照区块时间顺序解析
func (s *SwapCount) ParserDate() error {
	for {
		swapCount, err := model.QuerySwapCount(context.TODO(), model.SwapAddress(s.SwapAccount))
		if err != nil && !errors.Is(err, errors.RecordNotFound) {
			return errors.Wrap(err)
		}

		if swapCount != nil {
			s.ID = swapCount.LastSwapTransactionID
		}

		filters := []model.Filter{
			model.NewFilter("id <= ?", s.LastTransactionID),
			model.SwapAddress(s.SwapAccount),
			model.OrderFilter("id asc"),
			model.NewFilter("id > ?", s.ID),
		}

		swapTransactions, err := model.QuerySwapTransactions(context.TODO(), 100, 0, filters...)
		if err != nil {
			logger.Error("get single transaction err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if len(swapTransactions) == 0 {
			logger.Info(fmt.Sprintf("parse swap, swap address: %s , current id is %d, target id is %d", s.SwapAccount, s.ID, s.LastTransactionID))
			break
		}

		for _, transaction := range swapTransactions {
			s.ID = transaction.ID

			tx := parse.NewTx(transaction.TxData)
			err = tx.ParseTxToSwap()
			if err != nil {
				if errors.Is(err, errors.RecordNotFound) {
					continue
				}
				logger.Error("sync transaction id err", logger.Errorv(err))
				return errors.Wrap(err)
			}

			s.SwapRecords = tx.SwapRecords
			s.BlockDate = transaction.BlockTime

			if err = s.WriteToDB(transaction); err != nil {
				return errors.Wrap(err)
			}
		}

		// 更新处理数据的位置
		if err = model.UpdateSwapCountBySwapAccount(context.TODO(), s.SwapAccount, map[string]interface{}{"last_swap_transaction_id": s.ID}); err != nil {
			return errors.Wrap(err)
		}

		logger.Info(fmt.Sprintf("parse swap, swap address: %s , current id is %d, target id is %d", s.SwapAccount, s.ID, s.LastTransactionID))

	}

	return nil
}

func (s *SwapCount) WriteToDB(tx *domain.SwapTransaction) error {
	var err error
	trans := func(ctx context.Context) error {
		for _, swapRecord := range s.SwapRecords {
			// 仅当前swapAccount  可以插入
			if swapRecord.SwapConfig.SwapAccount != s.SwapAccount {
				continue
			}

			if err = s.updateSwapCount(ctx, swapRecord); err != nil {
				return errors.Wrap(err)
			}

			var (
				tokenAVolume      decimal.Decimal
				tokenBVolume      decimal.Decimal
				tokenAQuoteVolume decimal.Decimal
				tokenBQuoteVolume decimal.Decimal
			)
			if swapRecord.Direction == 0 {
				tokenAVolume = swapRecord.TokenCount.TokenAVolume
				tokenBQuoteVolume = swapRecord.TokenCount.TokenBVolume
			} else {
				tokenBVolume = swapRecord.TokenCount.TokenBVolume
				tokenAQuoteVolume = swapRecord.TokenCount.TokenAVolume
			}

			swapCountKLine := &domain.SwapCountKLine{
				LastSwapTransactionID: s.ID,
				SwapAddress:           swapRecord.SwapConfig.SwapAccount,
				TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
				TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
				TokenAVolume:          tokenAVolume,
				TokenBVolume:          tokenBVolume,
				TokenAQuoteVolume:     tokenAQuoteVolume,
				TokenBQuoteVolume:     tokenBQuoteVolume,
				TokenABalance:         swapRecord.TokenCount.TokenABalance,
				TokenBBalance:         swapRecord.TokenCount.TokenBBalance,
				DateType:              domain.DateMin,
				Open:                  swapRecord.Price,
				High:                  swapRecord.Price,
				Low:                   swapRecord.Price,
				Avg:                   swapRecord.Price,
				Settle:                swapRecord.Price,
				Date:                  s.BlockDate,
				TxNum:                 1,
				TokenAUSD:             tx.TokenAUSD,
				TokenBUSD:             tx.TokenBUSD,
				TokenASymbol:          swapRecord.SwapConfig.TokenA.Symbol,
				TokenBSymbol:          swapRecord.SwapConfig.TokenB.Symbol,
				TvlInUsd:              swapRecord.TokenCount.TokenABalance.Mul(tx.TokenAUSD).Add(swapRecord.TokenCount.TokenBBalance.Mul(tx.TokenBUSD)),
				VolInUsd:              tokenAVolume.Mul(tx.TokenAUSD).Abs().Add(tokenBVolume.Mul(tx.TokenBUSD)).Abs(),
			}

			newKline := kline.NewKline(s.BlockDate)
			for _, t := range newKline.Types {
				swapCountKLine.DateType = t.DateType
				if err = UpdateSwapCountKline(ctx, swapCountKLine, t); err != nil {
					return errors.Wrap(err)
				}
			}

		}
		return nil
	}

	if err = model.Transaction(context.TODO(), trans); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func UpdateSwapCountKline(ctx context.Context, swapCountKLine *domain.SwapCountKLine, t *kline.Type) error {

	currentSwapCountKLine, err := model.QuerySwapCountKLine(ctx,
		model.NewFilter("swap_address = ?", swapCountKLine.SwapAddress),
		model.NewFilter("date = ?", t.Date),
		model.NewFilter("date_type = ?", t.DateType))

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if currentSwapCountKLine != nil {
		if currentSwapCountKLine.High.GreaterThan(swapCountKLine.High) {
			swapCountKLine.High = currentSwapCountKLine.High
		}
		if currentSwapCountKLine.Low.LessThan(swapCountKLine.Low) {
			swapCountKLine.Low = currentSwapCountKLine.Low
		}
	}

	if t.DateType != domain.DateMin {
		innerAvg, err := t.CalculateAvg(func(endTime time.Time, avgList *[]*kline.InterTime) error {
			swapCountKLines, err := model.QuerySwapCountKLines(ctx, t.Interval, 0,
				model.NewFilter("date_type = ?", t.BeforeIntervalDateType),
				model.SwapAddress(swapCountKLine.SwapAddress),
				model.NewFilter("date < ?", endTime),
				model.OrderFilter("date desc"),
			)

			if err != nil {
				return errors.Wrap(err)
			}

			// 减少for 循环
			swapCountKLineMap := make(map[int64]*domain.SwapCountKLine, len(swapCountKLines))
			for index := range swapCountKLines {
				swapCountKLineMap[swapCountKLines[index].Date.Unix()] = swapCountKLines[index]
			}

			// 找到第一个数据
			lastAvg := &domain.SwapCountKLine{}
			for index := range swapCountKLines {
				if swapCountKLines[len(swapCountKLines)-index-1].Date.After((*avgList)[0].Date) {
					break
				}
				lastAvg = swapCountKLines[len(swapCountKLines)-index-1]
			}

			for index, avg := range *avgList {
				lastSwapCountKLine, ok := swapCountKLineMap[avg.Date.Unix()]
				if ok {
					lastAvg = lastSwapCountKLine
					(*avgList)[index].Avg = lastSwapCountKLine.Avg
					(*avgList)[index].TokenAUSD = lastSwapCountKLine.TokenAUSD
					(*avgList)[index].TokenBUSD = lastSwapCountKLine.TokenBUSD
				} else {
					(*avgList)[index].Avg = lastAvg.Settle // 上一个周期的结束值用作空缺周期的平均值
					(*avgList)[index].TokenAUSD = lastAvg.TokenAUSD
					(*avgList)[index].TokenBUSD = lastAvg.TokenBUSD
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err)
		}

		swapCountKLine.Avg = innerAvg.Avg
		swapCountKLine.TokenAUSD = innerAvg.TokenAUSD
		swapCountKLine.TokenBUSD = innerAvg.TokenBUSD
	}

	_, err = model.UpsertSwapCountKLine(ctx, swapCountKLine, t.Date)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (m *KLineTyp) updateKline(ctx context.Context, swapCountKLine *domain.SwapCountKLine) error {
	swapCountKLine.Date = m.GetDate()
	swapCountKLine.DateType = m.DateType
	currentSwapCountKLine, err := model.QuerySwapCountKLine(ctx,
		model.NewFilter("swap_address = ?", swapCountKLine.SwapAddress),
		model.NewFilter("date = ?", swapCountKLine.Date),
		model.NewFilter("date_type = ?", swapCountKLine.DateType))

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if currentSwapCountKLine != nil {
		if currentSwapCountKLine.High.GreaterThan(swapCountKLine.High) {
			swapCountKLine.High = currentSwapCountKLine.High
		}
		if currentSwapCountKLine.Low.LessThan(swapCountKLine.Low) {
			swapCountKLine.Low = currentSwapCountKLine.Low
		}
	}

	if m.DateType != domain.DateMin {
		avg, err := m.calculateAvg(ctx, swapCountKLine.SwapAddress)
		if err != nil {
			return errors.Wrap(err)
		}

		swapCountKLine.Avg = avg
	}

	_, err = model.UpsertSwapCountKLine(ctx, swapCountKLine, swapCountKLine.Date)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// calculateAvg 按照上一个周期计算平均值，month除外（按照天计算）
func (m *KLineTyp) calculateAvg(ctx context.Context, swapAccount string) (decimal.Decimal, error) {
	type interTime struct {
		Date time.Time
		avg  *decimal.Decimal
	}

	var (
		count     = int32(0)
		sum       = decimal.Zero
		beginTime time.Time
		endTime   time.Time
	)

	avgList := make([]*interTime, m.Interval, m.Interval)
	if m.DateType != domain.DateMin {
		beginTime = *m.GetDate()
		if m.DateType == domain.DateMon {
			lastDateTime := m.Date.AddDate(0, 1, -m.Date.Day())
			endTime = time.Date(lastDateTime.Year(), lastDateTime.Month(), lastDateTime.Day(), 0, 0, 0, 0, m.Date.Location()).Add(time.Hour * 24)
		} else {
			endTime = beginTime.Add(m.InnerTimeInterval * time.Duration(m.Interval))
		}

		swapCountKLines, err := model.QuerySwapCountKLines(ctx, m.Interval, 0,
			model.NewFilter("date_type = ?", m.BeforeIntervalDateType),
			model.SwapAddress(swapAccount),
			model.NewFilter("date < ?", endTime),
			model.OrderFilter("date desc"),
		)

		if err != nil {
			return decimal.Zero, errors.Wrap(err)
		}

		for index := range avgList {
			avgList[index] = &interTime{
				Date: m.GetDate().Add(m.InnerTimeInterval * time.Duration(index)),
			}
		}

		// 减少for 循环
		swapCountKLineMap := make(map[int64]*domain.SwapCountKLine, len(swapCountKLines))
		for index := range swapCountKLines {
			swapCountKLineMap[swapCountKLines[index].Date.Unix()] = swapCountKLines[index]
		}

		// 找到第一个数据
		lastAvg := &domain.SwapCountKLine{}
		for index := range swapCountKLines {
			if swapCountKLines[len(swapCountKLines)-index-1].Date.After(avgList[0].Date) {
				break
			}
			lastAvg = swapCountKLines[len(swapCountKLines)-index-1]
		}

		for index, avg := range avgList {
			lastSwapCountKLine, ok := swapCountKLineMap[avg.Date.Unix()]
			if ok {
				lastAvg = lastSwapCountKLine
				avgList[index].avg = &lastSwapCountKLine.Avg
			} else {
				avgList[index].avg = &lastAvg.Settle // 上一个周期的结束值用作空缺周期的平均值
			}
		}

		// calculate avg
		for _, v := range avgList {
			if !v.avg.IsZero() {
				sum = sum.Add(*v.avg)
				count++
			}
		}
	} else {
		return decimal.Zero, errors.New("error date_type")
	}

	return sum.Div(decimal.NewFromInt32(count)), nil
}

func (s *SwapCount) updateSwapCount(ctx context.Context, swapRecord *parse.SwapRecord) error {
	swapCount := &domain.SwapCount{
		LastSwapTransactionID: s.ID,
		SwapAddress:           swapRecord.SwapConfig.SwapAccount,
		TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
		TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
		TokenAVolume:          swapRecord.TokenCount.TokenAVolume,
		TokenBVolume:          swapRecord.TokenCount.TokenBVolume,
		TokenABalance:         swapRecord.TokenCount.TokenABalance,
		TokenBBalance:         swapRecord.TokenCount.TokenBBalance,
	}

	_, err := model.UpsertSwapCount(ctx, swapCount)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
