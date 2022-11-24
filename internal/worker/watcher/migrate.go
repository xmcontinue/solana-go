package watcher

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
)

func migrateSingleSwapPairPriceKlineBySwapAddress(swapAddress string) error {
	pairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddressFilter(swapAddress))
	if err != nil {
		return errors.Wrap(err)
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

		syncMigrateID := priceKLines[len(priceKLines)-1].ID
		// 去掉主键
		for _, v := range priceKLines {
			v.ID = 0
		}

		trans := func(ctx context.Context) error {
			err = model.CreateSwapPairPriceKLine(ctx, priceKLines)
			if err != nil {
				return errors.Wrap(err)
			}

			err = model.UpdateSwapPairBase(ctx, map[string]interface{}{
				"pair_price_migrate_id": syncMigrateID,
			},
				model.SwapAddressFilter(swapAddress),
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
		return errors.Wrap(err)
	}

	beginID := pairBase.MigrateID

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapAddress),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		// 只在原始表查询数据
		transactions, err := model.QuerySwapTransactionsV2InBaseTable(context.Background(), 1000, 0, filters...)
		if err != nil {
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

	// swap price 数据迁移
	err = migrateSwapPairPriceKline()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
