package watcher

import (
	"context"
	"strconv"
	"sync"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

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
		if config.Version != "v1" {
			// 只迁移v1
			continue
		}

		err := migrateSwapTransactionBySwapAddress(config.SwapAccount)
		if err != nil {
			return errors.Wrap(err)
		}

	}

	return nil
}

func migrateSingleSwapPairPriceKlineBySwapAddress(wg *sync.WaitGroup, limitChan chan struct{}, swapConfig *domain.SwapConfig) error {
	defer func() {
		<-limitChan
		wg.Done()
	}()

	var beginID int64
	pairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddressFilter(swapConfig.SwapAccount))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err)
		}

		err = model.CreateSwapPairBase(context.Background(), &domain.SwapPairBaseSharding{
			SwapAddress:      swapConfig.SwapAccount,
			TokenAAddress:    swapConfig.TokenA.SwapTokenAccount,
			TokenBAddress:    swapConfig.TokenB.SwapTokenAccount,
			IsSync:           false,
			SyncUtilFinished: true,
		})

		if err != nil {
			return errors.Wrap(err)
		}

	} else {
		if pairBase.IsPriceMigrateFinished == true {
			return nil
		}
		beginID = pairBase.PairPriceMigrateID
	}

	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapConfig.SwapAccount),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
		}
		// 只在原始表查询数据
		priceKLines, err := model.QuerySwapPairPriceKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(priceKLines) == 0 {
			break
		}
		logger.Info("migrate "+swapConfig.SwapAccount, logger.String(strconv.Itoa(int(beginID)), "a"))
		syncMigrateID := priceKLines[len(priceKLines)-1].ID

		for _, v := range priceKLines {
			v.ID = 0 // 去掉主键
		}

		err = model.UpdateSwapPairBase(context.Background(), map[string]interface{}{
			"pair_price_migrate_id": syncMigrateID,
		},
			model.SwapAddressFilter(swapConfig.SwapAccount),
			model.NewFilter("pair_price_migrate_id < ?", syncMigrateID),
		)

		if err != nil {
			return errors.Wrap(err)
		}

		// 插入数据
		err = model.CreateSwapPairPriceKLine(context.Background(), priceKLines)
		if err != nil {
			return errors.Wrap(err)
		}
		//
		//trans := func(ctx context.Context) error {
		//	err = model.UpdateSwapPairBase(ctx, map[string]interface{}{
		//		"pair_price_migrate_id": syncMigrateID,
		//	},
		//		model.SwapAddressFilter(swapConfig.SwapAccount),
		//	)
		//
		//	if err != nil {
		//		return errors.Wrap(err)
		//	}
		//
		//	// 插入数据
		//	err = model.CreateSwapPairPriceKLine(ctx, priceKLines)
		//	if err != nil {
		//		return errors.Wrap(err)
		//	}
		//
		//	return nil
		//}
		//
		//err = model.Transaction(context.Background(), trans)
		//if err != nil {
		//	return errors.Wrap(err)
		//}

		beginID = syncMigrateID

	}

	//同步完了就更新状态
	err = model.UpdateSwapPairBase(context.Background(), map[string]interface{}{
		"is_price_migrate_finished": true,
	},
		model.SwapAddressFilter(swapConfig.SwapAccount),
	)
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("finished migrate "+swapConfig.SwapAccount, logger.String(strconv.Itoa(int(beginID)), "a"))

	return nil
}

func migrateSwapPairPriceKline() error {
	configs := sol.SwapConfigList()

	limitChan := make(chan struct{}, len(configs))

	wg := &sync.WaitGroup{}
	wg.Add(len(configs))

	for i := range configs {
		limitChan <- struct{}{}
		go migrateSingleSwapPairPriceKlineBySwapAddress(wg, limitChan, configs[i])

	}
	wg.Wait()

	return nil
}

func migrate() error {

	// 原始tx 数据迁移
	//err = migrateSwapTransaction()
	//if err != nil {
	//	return errors.Wrap(err)
	//}
	logger.Info("migrate migrateSwapTransaction done")
	// swap price 数据迁移
	_ = migrateSwapPairPriceKline()

	logger.Info("migrate done")
	// 等迁移完成后才能解析其他数据

	return nil
}
