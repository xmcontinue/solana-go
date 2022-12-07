package process

import (
	"context"
	"strconv"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// swap count kline 数据迁移

func migrateSwapCountKlineBySwapAddress(swapConfig *domain.SwapConfig) error {
	beginID := int64(0)
	swapCount, err := model.QuerySwapCountMigrate(context.Background(), model.SwapAddressFilter(swapConfig.SwapAccount))
	if err != nil {
		err = model.CreateSwapCountMigrate(context.Background(), &domain.SwapCountMigrate{
			SwapAddress: swapConfig.SwapAccount,
		})
		if err != nil {
			return errors.Wrap(err)
		}
	} else {
		beginID = swapCount.MigrateSwapContKLineID
	}

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapConfig.SwapAccount),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		logger.Info(swapConfig.SwapAccount, logger.String("begin", strconv.FormatInt(beginID, 10)))
		// 只在原始表查询数据
		swapCountKLines, err := model.QuerySwapCountKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return errors.Wrap(err)
		}

		if len(swapCountKLines) == 0 {
			break
		}

		logger.Info(swapConfig.SwapAccount, logger.String("begin", strconv.FormatInt(beginID, 10)))
		syncMigrateID := swapCountKLines[len(swapCountKLines)-1].ID

		for _, v := range swapCountKLines {
			v.ID = 0
		}

		trans := func(ctx context.Context) error {
			logger.Info("aaaaa", logger.String(swapConfig.SwapAccount, "1"))
			err = model.UpdateSwapCountMigrate(ctx, map[string]interface{}{
				"migrate_swap_cont_k_line_id": syncMigrateID,
			},
				model.SwapAddressFilter(swapConfig.SwapAccount),
			)
			if err != nil {
				return errors.Wrap(err)
			}
			logger.Info("aaaaa", logger.String(swapConfig.SwapAccount, "2"))
			err = model.CreateSwapCountKLine(ctx, swapCountKLines)
			if err != nil {
				return errors.Wrap(err)
			}
			logger.Info("aaaaa", logger.String(swapConfig.SwapAccount, "3"))
			return nil
		}

		if err = model.Transaction(context.TODO(), trans); err != nil {
			logger.Error("transaction error", logger.Errorv(err))
			return errors.Wrap(err)
		}

		beginID = syncMigrateID

	}

	return nil
}

func migrateSwapCountKline() error {
	configs := sol.SwapConfigListV1()

	for i := range configs {

		err := migrateSwapCountKlineBySwapAddress(configs[i])
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil

}

func migrate() error {
	if !model.ISSharding {
		return nil
	}
	var err error

	logger.Info("migrate migrateSwapPairPriceKline done")
	// 等迁移完成后才能解析其他数据
	err = migrateSwapCountKline()
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("migrate migrateSwapCountKline done")
	return nil
}
