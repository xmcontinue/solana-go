package process

import (
	"context"
	"fmt"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
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
		err := processTransactionUserCount("v2", swap.SwapAccount)
		if err != nil {
			logger.Error("sync block time error", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}
	return nil
}

func processTransactionUserCount(version, swapAddress string) error {
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

	var userAddress []*model.UserAddressId
	for {

		if version == "v1" {
			userAddress, err = model.QuerySwapTransactionsUserAddress(ctx, 1000, 0, model.OrderFilter("id asc"), model.NewFilter("id > ?", beginID))
			if err != nil {
				if errors.Is(err, errors.RecordNotFound) {
					break
				}
				return errors.Wrap(err)
			}
		} else {
			userAddress, err = model.QuerySwapTransactionsUserAddressV2(ctx, 1000, 0, model.OrderFilter("id asc"), model.NewFilter("id > ?", beginID), model.SwapAddressFilter(swapAddress))
			if err != nil {
				if errors.Is(err, errors.RecordNotFound) {
					break
				}
				return errors.Wrap(err)
			}
		}

		if len(userAddress) == 0 {
			return nil
		}

		for _, v := range userAddress {
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

		beginID = userAddress[len(userAddress)-1].ID

	}

	return nil
}

func parserTransactionUserCount() error {
	// 先解析v1
	err := processTransactionUserCount("v1", "v1")
	if err != nil {
		return errors.Wrap(err)
	}

	err = parserTransactionV2UserCount()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
func yanzheng() error {
	transaction, err := model.QuerySwapTransactionsV2(context.Background(), 1, 0, model.SwapAddressFilter("BsgTBhUa9Nrs8GNjBoPDxgk4MzjUWVjtaRXAGZkFwxWa"),
		model.NewFilter("signature = ?", "rQDK3ThejjehLvUb8sN6qRFEzjyqLzVPeZMZiFZq2F4JAAnZknEE2ANmx1vgzSfSAbP28NmHXoh5Ng4QLnHV4bh"))
	if err != nil {
		logger.Error("fix", logger.Errorv(err))
		return errors.Wrap(err)
	}

	tx := parse.NewTxV2()

	err = tx.ParseSwapV2(transaction[0].Msg)
	if err != nil {
		if errors.Is(err, errors.RecordNotFound) {
			logger.Error("fix", logger.Errorv(err))
		}
		logger.Error("sync transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	if len(tx.SwapRecords) == 0 {
		return nil
	}

	fmt.Println("AmountIn:", tx.SwapRecords[0].AmountIn.String(), "AmountOut", tx.SwapRecords[0].AmountOut.String())
	fmt.Println("方向：", tx.SwapRecords[0].Direction)
	return nil
}
