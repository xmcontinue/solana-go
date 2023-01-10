package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
)

// token a volume token b volume ,fee amount， vol for usd
func syncPrice(swapAccount string, t time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	var volForUsd, swapAVolForUsd, swapBVolForUsd, swapFeeVolForUsd decimal.Decimal

	for {
		transactions, err := model.QuerySwapTransactionsV2(context.Background(), 10000, 0, model.SwapAddressFilter(swapAccount), model.NewFilter("block_time > ?", t), model.OrderFilter("id asc"))
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				break
			}
			return decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, err
		}

		if len(transactions) != 0 {
			t = *transactions[len(transactions)-1].BlockTime
		}

		for _, transaction := range transactions {
			tx := parse.NewTxV2()

			err := tx.ParseSwapV2(transaction.Msg)
			if err != nil {
				if errors.Is(err, errors.RecordNotFound) {
					continue
				}
				logger.Error("sync transaction id err", logger.Errorv(err))
				return decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, errors.Wrap(err)
			}

			if len(tx.SwapRecords) == 0 {
				continue
			}

			for _, v := range tx.SwapRecords {
				var tempAVolForUsd, tempBVolForUsd, feeVolForUsd decimal.Decimal
				if v.Direction == 0 {
					tempAVolForUsd = v.AmountIn.Mul(transaction.TokenAUSD)
					tempBVolForUsd = v.AmountOut.Mul(transaction.TokenBUSD)
					feeVolForUsd = v.FeeAmount.Mul(transaction.TokenAUSD)
				} else {
					tempAVolForUsd = v.AmountOut.Mul(transaction.TokenAUSD)
					tempBVolForUsd = v.AmountIn.Mul(transaction.TokenBUSD)
					feeVolForUsd = v.FeeAmount.Mul(transaction.TokenBUSD)
				}

				volForUsd = volForUsd.Add(tempAVolForUsd)
				swapAVolForUsd = swapAVolForUsd.Add(tempAVolForUsd)
				swapBVolForUsd = swapBVolForUsd.Add(tempBVolForUsd)
				swapFeeVolForUsd = swapFeeVolForUsd.Add(feeVolForUsd)

			}

		}
	}

	return volForUsd, swapAVolForUsd, swapBVolForUsd, swapFeeVolForUsd, nil
}
