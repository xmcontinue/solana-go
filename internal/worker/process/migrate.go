package process

import (
	"context"
	"strconv"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
)

// swap count kline 数据迁移

func migrateSwapCountKline1(swapAddress string) error {
	beginID := int64(0)

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapAddress),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		// 只在原始表查询数据
		swapCountKLines, err := model.QuerySwapCountKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapCountKLines) == 0 {
			break
		}

		for _, v := range swapCountKLines {
			v.ID = 0
		}
		logger.Info(swapAddress, logger.String("begin", strconv.FormatInt(beginID, 10)))
		syncMigrateID := swapCountKLines[len(swapCountKLines)-1].ID

		trans := func(ctx context.Context) error {
			logger.Info("aaaaa", logger.String(swapAddress, "1"))
			err = model.UpdateSwapCount(ctx, map[string]interface{}{
				"migrate_swap_cont_k_line_id": syncMigrateID,
			},
				model.SwapAddressFilter(swapAddress),
			)
			if err != nil {
				return errors.Wrap(err)
			}
			logger.Info("aaaaa", logger.String(swapAddress, "2"))
			err = model.CreateSwapCountKLine(ctx, swapCountKLines)
			if err != nil {
				return errors.Wrap(err)
			}
			logger.Info("aaaaa", logger.String(swapAddress, "3"))
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
	configs := sol.SwapConfigList()

	for _, v := range configs {
		if v.Version != "v1" {
			continue
		}
		err := migrateSwapCountKline1(v.SwapAccount)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil

}

func migrate() error {
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
