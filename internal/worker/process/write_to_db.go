package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

func createSwapCountKLine(writeTyp *WriteTyp, tokenAUSD, tokenBUSD decimal.Decimal) *domain.SwapCountKLine {
	var swapCountKLine *domain.SwapCountKLine
	for _, swapRecord := range writeTyp.swapRecords {
		// 仅当前swapAccount  可以插入
		if swapRecord.GetSwapConfig().SwapAccount != writeTyp.SwapAccount {
			continue
		}
		var (
			tokenAVolume         decimal.Decimal
			tokenBVolume         decimal.Decimal
			tokenAQuoteVolume    decimal.Decimal
			tokenBQuoteVolume    decimal.Decimal
			TokenARefAmount      decimal.Decimal
			TokenAFeeAmount      decimal.Decimal
			TokenAProtocolAmount decimal.Decimal
			TokenBRefAmount      decimal.Decimal
			TokenBFeeAmount      decimal.Decimal
			TokenBProtocolAmount decimal.Decimal
		)

		if swapRecord.GetDirectionByVersion() == 0 {
			tokenAVolume = swapRecord.GetTokenAVolume()
			tokenBQuoteVolume = swapRecord.GetTokenBVolume()

			//TokenARefAmount = swapRecord.GetTokenARefAmount()
			//TokenAFeeAmount = swapRecord.GetTokenAFeeAmount()
			//TokenAProtocolAmount = swapRecord.GetTokenAProtocolAmount()
			// 修改
			TokenARefAmount = swapRecord.GetTokenARefAmount()
			TokenAFeeAmount = swapRecord.GetTokenAFeeAmount()
			TokenAProtocolAmount = swapRecord.GetTokenAProtocolAmount()

		} else {
			tokenBVolume = swapRecord.GetTokenBVolume()
			tokenAQuoteVolume = swapRecord.GetTokenAVolume()

			TokenBRefAmount = swapRecord.GetTokenBRefAmount()
			TokenBFeeAmount = swapRecord.GetTokenBFeeAmount()
			TokenBProtocolAmount = swapRecord.GetTokenBProtocolAmount()
		}

		swapCountKLine = &domain.SwapCountKLine{
			LastSwapTransactionID:    writeTyp.ID,
			SwapAddress:              swapRecord.GetSwapConfig().SwapAccount,
			TokenAAddress:            swapRecord.GetSwapConfig().TokenA.SwapTokenAccount,
			TokenBAddress:            swapRecord.GetSwapConfig().TokenB.SwapTokenAccount,
			TokenAVolume:             tokenAVolume,
			TokenBVolume:             tokenBVolume,
			TokenAQuoteVolume:        tokenAQuoteVolume,
			TokenBQuoteVolume:        tokenBQuoteVolume,
			TokenABalance:            swapRecord.GetTokenABalance(),
			TokenBBalance:            swapRecord.GetTokenBBalance(),
			TokenARefAmount:          TokenARefAmount,
			TokenAFeeAmount:          TokenAFeeAmount,
			TokenAProtocolAmount:     TokenAProtocolAmount,
			TokenBRefAmount:          TokenBRefAmount,
			TokenBFeeAmount:          TokenBFeeAmount,
			TokenBProtocolAmount:     TokenBProtocolAmount,
			DateType:                 domain.DateMin,
			Open:                     swapRecord.GetPrice(),
			High:                     swapRecord.GetPrice(),
			Low:                      swapRecord.GetPrice(),
			Avg:                      swapRecord.GetPrice(),
			Settle:                   swapRecord.GetPrice(),
			Date:                     writeTyp.BlockDate,
			TxNum:                    1,
			TokenAUSD:                tokenAUSD,
			TokenBUSD:                tokenBUSD,
			TokenASymbol:             swapRecord.GetSwapConfig().TokenA.Symbol,
			TokenBSymbol:             swapRecord.GetSwapConfig().TokenB.Symbol,
			TvlInUsd:                 swapRecord.GetTokenABalance().Mul(tokenAUSD).Add(swapRecord.GetTokenBBalance().Mul(tokenBUSD)),
			VolInUsd:                 tokenAVolume.Mul(tokenAUSD).Abs().Add(tokenBVolume.Mul(tokenBUSD)).Abs(),
			MaxBlockTimeWithDateType: writeTyp.BlockDate,
		}
	}

	return swapCountKLine
}

// updateSwapCountKline 按照时间类型更新表
func updateSwapCountKline(ctx context.Context, swapCountKLine *domain.SwapCountKLine, t *kline.Type) error {

	currentSwapCountKLine, err := model.QuerySwapCountKLine(ctx,
		model.NewFilter("swap_address = ?", swapCountKLine.SwapAddress),
		model.NewFilter("date = ?", swapCountKLine.Date),
		model.NewFilter("date_type = ?", swapCountKLine.DateType))

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}
	logger.Info("测试", logger.String(swapCountKLine.SwapAddress, "8"))
	var tokenABalance, tokenBBalance decimal.Decimal
	var maxBlockTimeWithDateType *time.Time
	if currentSwapCountKLine != nil {
		if currentSwapCountKLine.High.GreaterThan(swapCountKLine.High) {
			swapCountKLine.High = currentSwapCountKLine.High
		}
		if currentSwapCountKLine.Low.LessThan(swapCountKLine.Low) {
			swapCountKLine.Low = currentSwapCountKLine.Low
		}

		if currentSwapCountKLine.MaxBlockTimeWithDateType != nil &&
			currentSwapCountKLine.MaxBlockTimeWithDateType.After(*swapCountKLine.MaxBlockTimeWithDateType) {
			tokenABalance = currentSwapCountKLine.TokenABalance
			tokenBBalance = currentSwapCountKLine.TokenBBalance
			maxBlockTimeWithDateType = currentSwapCountKLine.MaxBlockTimeWithDateType
		} else {
			tokenABalance = swapCountKLine.TokenABalance
			tokenBBalance = swapCountKLine.TokenBBalance
			maxBlockTimeWithDateType = swapCountKLine.MaxBlockTimeWithDateType
		}
	} else {
		tokenABalance = swapCountKLine.TokenABalance
		tokenBBalance = swapCountKLine.TokenBBalance
		maxBlockTimeWithDateType = swapCountKLine.MaxBlockTimeWithDateType
	}

	if swapCountKLine.DateType != domain.DateMin {
		innerAvg, err := t.CalculateAvg(func(endTime time.Time, avgList *[]*kline.InterTime) error {
			swapCountKLines, err := model.QuerySwapCountKLines(ctx, t.Interval, 0,
				model.NewFilter("date_type = ?", t.BeforeIntervalDateType),
				model.SwapAddressFilter(swapCountKLine.SwapAddress),
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
		logger.Info("测试", logger.String(swapCountKLine.SwapAddress, "9"))
		swapCountKLine.Avg = innerAvg.Avg
		swapCountKLine.TokenAUSD = innerAvg.TokenAUSD
		swapCountKLine.TokenBUSD = innerAvg.TokenBUSD
	}
	logger.Info("测试", logger.String(swapCountKLine.SwapAddress, "10"), logger.String(string(t.DateType), "aaaaa"))
	_, err = model.UpsertSwapCountKLine(ctx, swapCountKLine, tokenABalance, tokenBBalance, maxBlockTimeWithDateType)
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("测试", logger.String(swapCountKLine.SwapAddress, "20"))
	return nil
}

// getPriceInfo 获取价格信息
func getPriceInfo(ctx context.Context, date *time.Time, dateType domain.DateType, symbol string) (decimal.Decimal, error) {
	// 获取前一个时刻的价格
	var (
		TokenPriceInfo *domain.SwapTokenPriceKLine
		err            error
	)
	TokenPriceInfo, err = model.QuerySwapTokenPriceKLine(
		ctx,
		model.NewFilter("symbol = ?", symbol),
		model.NewFilter("date_type = ?", dateType),
		model.NewFilter("date <= ?", date),
		model.OrderFilter("date desc"),
	)
	if err != nil {
		// 获取后一个时刻的价格
		TokenPriceInfo, err = model.QuerySwapTokenPriceKLine(
			ctx,
			model.NewFilter("symbol = ?", symbol),
			model.NewFilter("date_type = ?", dateType),
			model.NewFilter("date > ?", date),
			model.OrderFilter("date asc"),
		)
		if err != nil {
			return decimal.Zero, errors.Wrap(err)
		}
	}
	return TokenPriceInfo.Avg, nil
}

func priceToSwapKLineHandle(ctx context.Context, swapInfo *domain.SwapCountKLine) (decimal.Decimal, decimal.Decimal, error) {
	var (
		tokenAPrice decimal.Decimal
		tokenBPrice decimal.Decimal
		err         error
	)
	// 查找tokenA,tokenB价格
	tokenAPrice, err = getPriceInfo(ctx, swapInfo.Date, swapInfo.DateType, swapInfo.TokenASymbol)
	if err != nil {
		return tokenAPrice, tokenBPrice, errors.Wrap(err)
	}
	tokenBPrice, err = getPriceInfo(ctx, swapInfo.Date, swapInfo.DateType, swapInfo.TokenBSymbol)
	if err != nil {
		return tokenAPrice, tokenBPrice, errors.Wrap(err)
	}

	return tokenAPrice, tokenBPrice, nil
}

// writeSwapRecordToDB 只存储swap tx 数据
func writeSwapRecordToDB(writeTyp *WriteTyp, tokenAUSD, tokenBUSD decimal.Decimal) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Recovered in f", err)
		}
	}()
	logger.Info("测试", logger.String(writeTyp.SwapAccount, "5"))
	var err error
	trans := func(ctx context.Context) error {
		swapCountKLine := createSwapCountKLine(writeTyp, tokenAUSD, tokenBUSD)

		if swapCountKLine == nil {
			return nil
		}
		logger.Info("测试", logger.String(writeTyp.SwapAccount, "6"))
		if _, err = model.UpsertSwapCount(ctx, &domain.SwapCount{
			LastSwapTransactionID: swapCountKLine.LastSwapTransactionID,
			SwapAddress:           swapCountKLine.SwapAddress,
			TokenAAddress:         swapCountKLine.TokenAAddress,
			TokenBAddress:         swapCountKLine.TokenBAddress,
			TokenAVolume:          swapCountKLine.TokenAVolume,
			TokenBVolume:          swapCountKLine.TokenBVolume,
			TokenABalance:         swapCountKLine.TokenABalance,
			TokenBBalance:         swapCountKLine.TokenBBalance,
		}); err != nil {
			return errors.Wrap(err)
		}

		newKline := kline.NewKline(writeTyp.BlockDate)
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
			logger.Info("测试", logger.String(writeTyp.SwapAccount, "7"))
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
	logger.Info("测试", logger.String(writeTyp.SwapAccount, "8"))
	return nil
}

func createSwapRecords(writeTyp *WriteTyp, userCountKLine *domain.UserCountKLine) {
	for _, swapRecord := range writeTyp.swapRecords {
		// 仅当前swapAccount  可以插入
		if swapRecord.GetSwapConfig().SwapAccount != writeTyp.SwapAccount {
			continue
		}

		var (
			tokenAVolume      decimal.Decimal
			tokenBVolume      decimal.Decimal
			tokenAQuoteVolume decimal.Decimal
			tokenBQuoteVolume decimal.Decimal
		)
		if swapRecord.GetDirectionByVersion() == 0 {
			tokenAVolume = swapRecord.GetTokenAVolume()
			tokenBQuoteVolume = swapRecord.GetTokenBVolume()
		} else {
			tokenBVolume = swapRecord.GetTokenBVolume()
			tokenAQuoteVolume = swapRecord.GetTokenAVolume()
		}

		userCountKLine.UserAddress = swapRecord.GetUserOwnerAccount()
		userCountKLine.SwapAddress = swapRecord.GetSwapConfig().SwapAccount
		userCountKLine.LastSwapTransactionID = writeTyp.ID
		userCountKLine.TokenAAddress = swapRecord.GetSwapConfig().TokenA.SwapTokenAccount
		userCountKLine.TokenBAddress = swapRecord.GetSwapConfig().TokenB.SwapTokenAccount
		userCountKLine.Date = writeTyp.BlockDate
		userCountKLine.TxNum += 1
		userCountKLine.TokenASymbol = swapRecord.GetSwapConfig().TokenA.Symbol
		userCountKLine.TokenBSymbol = swapRecord.GetSwapConfig().TokenB.Symbol

		userCountKLine.UserTokenAVolume = userCountKLine.UserTokenAVolume.Add(tokenAVolume)
		userCountKLine.UserTokenBVolume = userCountKLine.UserTokenBVolume.Add(tokenBVolume)
		userCountKLine.TokenAQuoteVolume = userCountKLine.TokenAQuoteVolume.Add(tokenAQuoteVolume)
		userCountKLine.TokenBQuoteVolume = userCountKLine.TokenBQuoteVolume.Add(tokenBQuoteVolume)
	}
}

func createLiquidity(writeTyp *WriteTyp, userCountKLine *domain.UserCountKLine) {
	for _, liquidity := range writeTyp.liquidityRecords {
		// 仅当前swapAccount  可以插入
		if liquidity.GetSwapConfig().SwapAccount != writeTyp.SwapAccount {
			continue
		}

		userCountKLine.UserAddress = liquidity.GetUserOwnerAccount()
		userCountKLine.SwapAddress = liquidity.GetSwapConfig().SwapAccount
		userCountKLine.LastSwapTransactionID = writeTyp.ID
		userCountKLine.Date = writeTyp.BlockDate
		userCountKLine.TxNum += 1
		userCountKLine.TokenAAddress = liquidity.GetSwapConfig().TokenA.SwapTokenAccount
		userCountKLine.TokenBAddress = liquidity.GetSwapConfig().TokenB.SwapTokenAccount
		userCountKLine.TokenASymbol = liquidity.GetSwapConfig().TokenA.Symbol
		userCountKLine.TokenBSymbol = liquidity.GetSwapConfig().TokenB.Symbol
		if liquidity.GetDirection() == 0 {
			userCountKLine.TokenAWithdrawLiquidityVolume = userCountKLine.TokenAWithdrawLiquidityVolume.Add(liquidity.GetTokenALiquidityVolume())
			userCountKLine.TokenBWithdrawLiquidityVolume = userCountKLine.TokenBWithdrawLiquidityVolume.Add(liquidity.GetTokenBLiquidityVolume())

		} else {
			userCountKLine.TokenADepositLiquidityVolume = userCountKLine.TokenADepositLiquidityVolume.Add(liquidity.GetTokenALiquidityVolume())
			userCountKLine.TokenBDepositLiquidityVolume = userCountKLine.TokenBDepositLiquidityVolume.Add(liquidity.GetTokenBLiquidityVolume())
		}
	}
}

func createClaim(writeTyp *WriteTyp, userCountKLine *domain.UserCountKLine) {

	for _, collectRecord := range writeTyp.claimRecords {
		// 仅当前swapAccount  可以插入
		if collectRecord.GetSwapConfig().SwapAccount != writeTyp.SwapAccount {
			continue
		}

		userCountKLine.UserAddress = collectRecord.GetUserOwnerAccount()
		userCountKLine.LastSwapTransactionID = writeTyp.ID
		userCountKLine.Date = writeTyp.BlockDate
		userCountKLine.TxNum += 1
		userCountKLine.SwapAddress = collectRecord.GetSwapConfig().SwapAccount
		userCountKLine.TokenAAddress = collectRecord.GetSwapConfig().TokenA.SwapTokenAccount
		userCountKLine.TokenBAddress = collectRecord.GetSwapConfig().TokenB.SwapTokenAccount
		userCountKLine.TokenASymbol = collectRecord.GetSwapConfig().TokenA.Symbol
		userCountKLine.TokenBSymbol = collectRecord.GetSwapConfig().TokenB.Symbol

		userCountKLine.TokenAClaimVolume = userCountKLine.TokenAClaimVolume.Add(collectRecord.GetTokenACollectVolume())
		userCountKLine.TokenBClaimVolume = userCountKLine.TokenBClaimVolume.Add(collectRecord.GetTokenBCollectVolume())
	}

}

func writeAllToDB(writeTyp *WriteTyp) error {
	var (
		err            error
		userCountKLine = &domain.UserCountKLine{}
	)
	trans := func(ctx context.Context) error {

		createSwapRecords(writeTyp, userCountKLine)

		createLiquidity(writeTyp, userCountKLine)

		createClaim(writeTyp, userCountKLine)

		newKline := kline.NewKline(writeTyp.BlockDate)
		for _, t := range newKline.Types {
			if t.DateType == domain.DateMin || t.DateType == domain.DateTwelfth || t.DateType == domain.DateQuarter || t.DateType == domain.DateHalfAnHour || t.DateType == domain.DateHour {
				continue
			}

			userCountKLine.DateType = t.DateType
			userCountKLine.Date = t.Date
			if _, err = model.UpsertUserSwapCountKLine(ctx, userCountKLine); err != nil {
				return errors.Wrap(err)
			}
		}
		return nil
	}

	if err = model.Transaction(context.TODO(), trans); err != nil {
		return errors.Wrap(err)
	}

	return nil
}
