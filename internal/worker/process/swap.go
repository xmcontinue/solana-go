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
)

// SwapAndUserCount 同步更新swap_counts表和user_swap_counts表
type SwapAndUserCount struct {
	ID                int64
	LastTransactionID int64
	SwapAccount       string
	SwapRecords       []*parse.SwapRecord
	BlockDate         *time.Time
	spec              string
}

// ParserDate 按照区块时间顺序解析
func (s *SwapAndUserCount) ParserDate() error {
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

func (s *SwapAndUserCount) WriteToDB(tx *domain.SwapTransaction) error {
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

			if err = s.userSwapCount(ctx, swapRecord, tx); err != nil {
				return errors.Wrap(err)
			}

			if err = s.userSwapCountDay(ctx, swapRecord, tx); err != nil {
				return errors.Wrap(err)
			}

			var (
				tokenAVolume decimal.Decimal
				tokenBVolume decimal.Decimal
			)
			if swapRecord.Direction == 0 {
				tokenAVolume = swapRecord.TokenCount.TokenAVolume
			} else {
				tokenBVolume = swapRecord.TokenCount.TokenBVolume
			}

			swapCountKLine := &domain.SwapCountKLine{
				LastSwapTransactionID: s.ID,
				SwapAddress:           swapRecord.SwapConfig.SwapAccount,
				TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
				TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
				TokenAVolume:          tokenAVolume,
				TokenBVolume:          tokenBVolume,
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

			for _, dateType := range []KLineTyp{DateMin, DateTwelfth, DateQuarter, DateHalfAnHour, DateHour, DateDay, DateWek, DateMon} {
				KLType := &KLineTyp{
					Date:                   tx.BlockTime,
					DateType:               dateType.DateType,
					BeforeIntervalDateType: dateType.BeforeIntervalDateType,
					Interval:               dateType.Interval,
					InnerTimeInterval:      dateType.InnerTimeInterval,
				}

				if err = KLType.updateKline(ctx, swapCountKLine); err != nil {
					return errors.Wrap(err)
				}
			}

			return nil
		}
		return nil
	}

	if err = model.Transaction(context.TODO(), trans); err != nil {
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
		avg, err := m.calculateAvg(ctx)
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

// 按照上一个周期计算平均值，month除外（按照天计算）
func (m *KLineTyp) calculateAvg(ctx context.Context) (decimal.Decimal, error) {
	type interTime struct {
		Date time.Time
		avg  decimal.Decimal
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
			endTime = time.Date(lastDateTime.Year(), lastDateTime.Month(), lastDateTime.Day(), 0, 0, 0, 0, m.Date.Location())
		} else {
			endTime = beginTime.Add(m.InnerTimeInterval * time.Duration(m.Interval))
		}

		swapCountKLines, err := model.QuerySwapCountKLines(ctx, m.Interval, 0,
			model.NewFilter("date_type = ?", m.BeforeIntervalDateType),
			model.NewFilter("date < ?", endTime))
		if err != nil {
			return decimal.Zero, errors.Wrap(err)
		}

		for index := range avgList {
			avgList[index] = &interTime{
				Date: m.GetDate().Add(m.InnerTimeInterval * time.Duration(index)),
			}
		}

		for _, v := range swapCountKLines {
			for _, avg := range avgList {
				if v.Date.Equal(avg.Date) || v.Date.Before(avg.Date) {
					avg.avg = v.Settle // 以上一个时间区间的结束值作为新的时间区间的平均值
				}
			}
		}

		// calculate avg
		for _, v := range avgList {
			if !v.avg.IsZero() {
				sum = sum.Add(v.avg)
				count++
			}
		}
	}
	return sum.Div(decimal.NewFromInt32(count)), nil
}

func (s *SwapAndUserCount) updateSwapCount(ctx context.Context, swapRecord *parse.SwapRecord) error {
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

// userSwapCount 写入user_counts 表
func (s *SwapAndUserCount) userSwapCount(ctx context.Context, swapRecord *parse.SwapRecord, tx *domain.SwapTransaction) error {
	userSwapCount := &domain.UserSwapCount{
		LastSwapTransactionID: s.ID,
		UserAddress:           swapRecord.UserTokenBAddress,
		SwapAddress:           swapRecord.SwapConfig.SwapAccount,
		TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
		TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
		UserTokenAVolume:      swapRecord.UserCount.TokenAVolume,
		UserTokenBVolume:      swapRecord.UserCount.TokenBVolume,
		UserTokenABalance:     swapRecord.UserCount.TokenABalance,
		UserTokenBBalance:     swapRecord.UserCount.TokenBBalance,
		TxNum:                 1,
		MaxTxVolume:           swapRecord.UserCount.TokenAVolume.Mul(tx.TokenAUSD),
		MinTxVolume:           swapRecord.UserCount.TokenAVolume.Mul(tx.TokenAUSD),
	}

	_, err := model.UpsertUserSwapCount(ctx, userSwapCount)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// userSwapCountDay 写入user_count_days 表
func (s *SwapAndUserCount) userSwapCountDay(ctx context.Context, swapRecord *parse.SwapRecord, tx *domain.SwapTransaction) error {
	userSwapCountDate := time.Date(s.BlockDate.Year(), s.BlockDate.Month(), s.BlockDate.Day(), 0, 0, 0, 0, s.BlockDate.Location())
	// 统计用户每日swap count
	userSwapCountDay := &domain.UserSwapCountDay{
		LastSwapTransactionID: s.ID,
		UserAddress:           swapRecord.UserOwnerAddress,
		SwapAddress:           swapRecord.SwapConfig.SwapAccount,
		TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
		TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
		UserTokenAVolume:      swapRecord.UserCount.TokenAVolume,
		UserTokenBVolume:      swapRecord.UserCount.TokenBVolume,
		UserTokenABalance:     swapRecord.UserCount.TokenABalance,
		UserTokenBBalance:     swapRecord.UserCount.TokenBBalance,
		TxNum:                 1,
		Date:                  &userSwapCountDate,
	}
	userSwapCountDays, total, err := model.QueryUserSwapCountDay(
		ctx,
		1,
		0,
		model.NewFilter("user_address = ?", swapRecord.UserOwnerAddress),
		model.NewFilter("swap_address = ?", swapRecord.SwapConfig.SwapAccount),
		model.NewFilter("date = ?", userSwapCountDate),
	)

	if err != nil {
		return errors.Wrap(err)
	}

	if total == 0 {
		userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume
		userSwapCountDay.MinTxVolume = userSwapCountDay.UserTokenAVolume
	} else {
		if userSwapCountDays[0].MaxTxVolume.LessThan(userSwapCountDay.UserTokenAVolume) {
			userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume.Mul(tx.TokenAUSD)
		}

		if userSwapCountDays[0].MinTxVolume.GreaterThan(userSwapCountDay.UserTokenAVolume) {
			userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume.Mul(tx.TokenAUSD)
		}
	}

	_, err = model.UpsertUserSwapCountDay(ctx, userSwapCountDay, s.BlockDate)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil

}
