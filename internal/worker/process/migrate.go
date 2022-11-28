package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

// swap count kline 数据迁移

func migrateSwapCountKline() error {
	beginID := int64(0)

	for {
		filters := []model.Filter{
			model.NewFilter("is_migrate = false"),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
			model.DateTypeFilter(domain.DateMin),
		}
		// 只在原始表查询数据
		swapCountKLines, err := model.QuerySwapCountKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapCountKLines) == 0 {
			break
		}

		syncMigrateID := swapCountKLines[len(swapCountKLines)-1].ID

		for _, swapCountKLine := range swapCountKLines {
			trans := func(ctx context.Context) error {
				if swapCountKLine == nil {
					return nil
				}

				if err = model.UpdateSwapCountKLinesInBaseTable(ctx,
					map[string]interface{}{
						"is_migrate": true,
					},
					model.IDFilter(swapCountKLine.ID),
					model.NewFilter("is_migrate = ?", false),
				); err != nil {
					return errors.Wrap(err)
				}

				// 去掉主键，由数据库自动生成
				swapCountKLine.ID = 0

				newKline := kline.NewKline(swapCountKLine.Date)
				for _, t := range newKline.Types {
					swapCountKLine.DateType = t.DateType
					swapCountKLine.Date = t.Date
					// 获取价格
					tokenAPrice, tokenBPrice, err := priceToSwapKLineHandle(ctx, swapCountKLine)
					if err != nil {
						return errors.Wrap(err)
					}
					swapCountKLine.TokenAUSDForContract = tokenAPrice
					swapCountKLine.TokenBUSDForContract = tokenBPrice

					if t.DateType == domain.DateMin {
						swapCountKLine.VolInUsdForContract = swapCountKLine.TokenAVolume.Mul(swapCountKLine.TokenAUSDForContract).Abs().Add(swapCountKLine.TokenBVolume.Mul(swapCountKLine.TokenBUSDForContract)).Abs()
					}

					if err = updateSwapCountKline(ctx, swapCountKLine, t); err != nil {
						return errors.Wrap(err)
					}
				}

				return nil
			}

			if err = model.Transaction(context.TODO(), trans); err != nil {
				logger.Error("transaction error", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}

		beginID = syncMigrateID

	}

	return nil
}

func migrateSingleSwapPairPriceKlineBySwapAddress(swapAddress string) error {
	pairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddressFilter(swapAddress))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err)
		}
		return nil
	}

	beginID := pairBase.PairPriceMigrateID

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapAddress),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		// 只在原始表查询数据
		priceKLines, err := model.QuerySwapPairPriceKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(priceKLines) == 0 {
			return nil
		}
		logger.Info("migrate "+swapAddress, logger.String(string(beginID), "a"))
		syncMigrateID := priceKLines[len(priceKLines)-1].ID

		trans := func(ctx context.Context) error {
			// 插入数据
			for _, v := range priceKLines {
				v.ID = 0 // 去掉主键
				_, err = model.UpsertSwapPairPriceKLine(ctx, v)
				if err != nil {
					return errors.Wrap(err)
				}
			}

			err = model.UpdateSwapPairBase(ctx, map[string]interface{}{
				"pair_price_migrate_id": syncMigrateID,
			},
				model.SwapAddressFilter(swapAddress),
				model.NewFilter("pair_price_migrate_id < ?", syncMigrateID),
			)

			if err != nil {
				return errors.Wrap(err)
			}

			return nil
		}

		err = model.Transaction(context.Background(), trans)
		if err != nil {
			return errors.Wrap(err)
		}

		beginID = syncMigrateID

	}
}

func updateSwapPairPrice(ctx context.Context, config *domain.SwapConfig, t *kline.Type, swapPairPriceKLine *domain.SwapPairPriceKLine) error {
	swapPriceKline, err := model.QuerySwapPairPriceKLine(ctx,
		model.SwapAddressFilter(config.SwapAccount),
		model.NewFilter("date = ?", t.Date),
		model.NewFilter("date_type = ?", t.DateType))

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if swapPriceKline != nil {
		if swapPriceKline.High.GreaterThan(swapPairPriceKLine.High) {
			swapPairPriceKLine.High = swapPriceKline.High
		}
		if swapPriceKline.Low.LessThan(swapPairPriceKLine.Low) {
			swapPairPriceKLine.Low = swapPriceKline.Low
		}
	}

	if t.DateType != domain.DateMin {

		InnerAvg, err := t.CalculateAvg(func(endTime time.Time, avgList *[]*kline.InterTime) error {
			swapCountKLines, err := model.QuerySwapPairPriceKLines(ctx, t.Interval, 0,
				model.SwapAddressFilter(config.SwapAccount),
				model.NewFilter("date_type = ?", t.BeforeIntervalDateType),
				model.NewFilter("date < ?", endTime),
				model.OrderFilter("date desc"),
			)

			if err != nil {
				return errors.Wrap(err)
			}

			for index := range swapCountKLines {
				for _, avg := range *avgList {
					if swapCountKLines[len(swapCountKLines)-1-index].Date.Equal(avg.Date) || swapCountKLines[len(swapCountKLines)-1-index].Date.Before(avg.Date) {
						avg.Avg = swapCountKLines[len(swapCountKLines)-1-index].Avg // 以上一个时间区间的平均值作为新的时间区间的平均值
					}
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err)
		}

		swapPairPriceKLine.Avg = InnerAvg.Avg

	}

	return nil
}

func migrateSwapPairPriceKline() error {
	configs := sol.SwapConfigList()
	for _, config := range configs {
		err := migrateSingleSwapPairPriceKlineBySwapAddress(config.SwapAccount)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

// 迁移swapTransaction
func migrateSwapTransactionBySwapAddress(swapAddress string) error {
	pairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddressFilter(swapAddress))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err)
		}
		return nil
	}

	beginID := pairBase.MigrateID

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapAddress),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		// 只在原始表查询数据
		transactions, err := model.QuerySwapTransactionsV2InBaseTable(context.Background(), 100, 0, filters...)
		if err != nil && !errors.Is(err, errors.RecordNotFound) {

			return errors.Wrap(err)
		}

		if len(transactions) == 0 {
			break
		}

		syncMigrateID := transactions[len(transactions)-1].ID
		// 去掉主键
		for _, v := range transactions {
			v.ID = 0
		}

		trans := func(ctx context.Context) error {
			err = model.UpdateSwapPairBase(ctx, map[string]interface{}{
				"migrate_id": syncMigrateID,
			},
				model.SwapAddressFilter(swapAddress),
			)
			if err != nil {
				return errors.Wrap(err)
			}

			err = model.CreatedSwapTransactionsV2(ctx, transactions)
			if err != nil {
				return errors.Wrap(err)
			}

			return nil
		}

		err = model.Transaction(context.Background(), trans)
		if err != nil {
			return errors.Wrap(err)
		}

		beginID = syncMigrateID
	}
	return nil
}

func migrateSwapTransaction() error {
	configs := sol.SwapConfigList()
	for _, config := range configs {
		if config.Version == "v1" {
			// 只迁移v2
			continue
		}

		err := migrateSwapTransactionBySwapAddress(config.SwapAccount)
		if err != nil {
			return errors.Wrap(err)
		}

	}

	return nil
}

func migrate() error {
	var err error
	// 原始tx 数据迁移
	err = migrateSwapTransaction()
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("migrate migrateSwapTransaction done")
	// swap price 数据迁移
	err = migrateSwapPairPriceKline()
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("migrate migrateSwapPairPriceKline done")
	// 等迁移完成后才能解析其他数据
	err = migrateSwapCountKline()
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("migrate migrateSwapCountKline done")
	return nil
}
