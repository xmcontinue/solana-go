package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func parserTransactionV2UserCount() error {
	swapConfig := sol.SwapConfigList()
	for _, swap := range swapConfig {
		if swap.Version != "v2" {
			continue
		}

		//err := parserSingleTransactionV2UserCount(swap)
		err := processTransactionUserCount(swap.SwapAccount)
		if err != nil {
			logger.Error("sync block time error", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}
	return nil
}

func processTransactionUserCount(swapAddress string) error {
	ctx := context.Background()
	beginID := int64(0)
	userCount, err := model.QuerySwapUserCount(ctx, model.SwapAddressFilter(swapAddress))
	if err != nil {
		if !errors.Is(err, errors.RecordNotFound) {
			return errors.Wrap(err)
		}

		err = model.CreateSwapUserCount(ctx, &domain.SwapUserCount{
			SwapAddress: swapAddress,
			SyncUtilID:  0,
		})
		if err != nil {
			return errors.Wrap(err)
		}
	} else {
		beginID = userCount.SyncUtilID
	}

	for {
		transactions, err := model.QuerySwapTransactions(ctx, 1000, 0, model.OrderFilter("id asc"), model.NewFilter("id > ?", beginID))
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				break
			}
			return errors.Wrap(err)
		}

		if len(transactions) == 0 {
			return nil
		}

		for _, v := range transactions {
			trans := func(ctx context.Context) error {
				err = model.UpdateSwapUserCount(ctx, map[string]interface{}{
					"sync_util_id": v.ID,
				},
					model.SwapAddressFilter(swapAddress),
				)
				if err != nil {
					return errors.Wrap(err)
				}

				_, err = model.QueryTransActionUserCount(ctx, model.NewFilter("user_address = ?", v.UserAddress))
				if err == nil {
					return nil
				}

				if !errors.Is(err, errors.RecordNotFound) {
					return errors.Wrap(err)
				}

				// 创建新的user_address 地址
				err = model.CreateTransActionUserCount(ctx, &domain.TransActionUserCount{
					UserAddress: v.UserAddress,
				})

				if err != nil {
					return errors.Wrap(err)
				}

				return nil
			}

			err = model.Transaction(context.Background(), trans)
			if err != nil {
				return errors.Wrap(err)
			}

		}

		beginID = transactions[len(transactions)-1].ID

	}

	return nil
}

func parserTransactionUserCount() error {
	// 先解析v1
	err := processTransactionUserCount("v1")
	if err != nil {
		return errors.Wrap(err)
	}

	err = parserTransactionV2UserCount()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
