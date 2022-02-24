package process

import (
	"context"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"github.com/gagliardetto/solana-go/rpc"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type syncType string

var (
	LastSwapTransactionID syncType = "last:swap:transaction:id"
	// 如果有新增的表，则新增redis key ，用以判断当前表同步数据位置，且LastSwapTransactionID为截止id
)

func syncData() error {
	lastSwapTransactionID, err := redisClient.Get(context.TODO(), string(LastSwapTransactionID)).Int64()
	if err != nil && !redisClient.ErrIsNil(err) {
		logger.Error("get ")
		return errors.Wrap(err)
	}

	for {
		swapTransactions, total, err := model.QuerySwapTransactions(context.TODO(), 1, 0, model.NewFilter("id > ?", lastSwapTransactionID))
		if err != nil {
			logger.Error("get single transaction err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if total == 0 {
			return nil
		}

		if err = parserData(context.TODO(), swapTransactions[0]); err != nil {
			logger.Error("parser data err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		// 同步到redis
		lastSwapTransactionID = swapTransactions[len(swapTransactions)-1].ID
		err = redisClient.Set(context.TODO(), string(LastSwapTransactionID), lastSwapTransactionID, 0).Err()
		if err != nil {
			logger.Error("sync transaction id err", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}
}

func parserSwapTvlCount(swapTransaction *domain.SwapTransaction) *domain.SwapTvlCount {
	// 统计tvl
	swapTvlCount := &domain.SwapTvlCount{
		LastSwapTransactionID: swapTransaction.ID,
		SwapAddress:           swapTransaction.SwapAddress,
		TokenAAddress:         swapTransaction.TokenAAddress,
		TokenBAddress:         swapTransaction.TokenBAddress,
		TokenAVolume:          swapTransaction.TokenAVolume,
		TokenBVolume:          swapTransaction.TokenBVolume,
		TokenABalance:         swapTransaction.TokenABalance,
		TokenBBalance:         swapTransaction.TokenBBalance,
		Tvl:                   swapTransaction.TokenABalance.Add(swapTransaction.TokenBBalance), // todo 汇率转换
		Vol:                   swapTransaction.TokenAVolume,                                     // todo 汇率转换
	}

	return swapTvlCount
}

func parserSwapTvlCountDate(swapTransaction *domain.SwapTransaction, blockDate *time.Time) *domain.SwapTvlCountDay {
	return &domain.SwapTvlCountDay{
		LastSwapTransactionID: swapTransaction.ID,
		SwapAddress:           swapTransaction.SwapAddress,
		TokenAAddress:         swapTransaction.TokenAAddress,
		TokenBAddress:         swapTransaction.TokenBAddress,
		TokenAVolume:          swapTransaction.TokenAVolume,
		TokenBVolume:          swapTransaction.TokenBVolume,
		TokenABalance:         swapTransaction.TokenABalance,
		TokenBBalance:         swapTransaction.TokenBBalance,
		Tvl:                   swapTransaction.TokenABalance.Add(swapTransaction.TokenBBalance), // todo 汇率转换
		Vol:                   swapTransaction.TokenAVolume,                                     // todo 汇率转换
		Date:                  blockDate,
		TxNum:                 1,
	}

}

func parserData(ctx context.Context, swapTransaction *domain.SwapTransaction) error {
	// 统计tvl
	swapTvlCount := parserSwapTvlCount(swapTransaction)

	blockDate := time.Date(swapTransaction.BlockTime.Year(), swapTransaction.BlockTime.Month(), swapTransaction.BlockTime.Day(), 0, 0, 0, 0, time.UTC)
	// 统计 每日 tvl
	swapTvlCountDay := parserSwapTvlCountDate(swapTransaction, &blockDate)

	transaction := &rpc.TransactionWithMeta{}
	//err := json.Unmarshal(swapTransaction.TxData, transaction)
	//if err != nil {
	//	return errors.Wrap(err)
	//}
	// 统计用户swap
	userTokenADeltaVolumeDecimal, userTokenBDeltaVolumeDecimal, userTokenABalance, userTokenBBalance := getUserSwapVolumeAndBalance(transaction)

	userSwapCount := &domain.UserSwapCount{
		LastSwapTransactionID: swapTransaction.ID,
		UserAddress:           swapTransaction.UserAddress,
		SwapAddress:           swapTransaction.SwapAddress,
		TokenAAddress:         swapTransaction.TokenAAddress,
		TokenBAddress:         swapTransaction.TokenBAddress,
		UserTokenAVolume:      userTokenADeltaVolumeDecimal,
		UserTokenBVolume:      userTokenBDeltaVolumeDecimal,
		UserTokenABalance:     userTokenABalance,
		UserTokenBBalance:     userTokenBBalance,
		TxNum:                 1,
	}

	// 统计用户每日swap count

	userSwapCountDay := &domain.UserSwapCountDay{
		LastSwapTransactionID: swapTransaction.ID,
		UserAddress:           swapTransaction.UserAddress,
		SwapAddress:           swapTransaction.SwapAddress,
		TokenAAddress:         swapTransaction.TokenAAddress,
		TokenBAddress:         swapTransaction.TokenBAddress,
		UserTokenAVolume:      userTokenADeltaVolumeDecimal,
		UserTokenBVolume:      userTokenBDeltaVolumeDecimal,
		UserTokenABalance:     userTokenABalance,
		UserTokenBBalance:     userTokenBBalance,
		TxNum:                 1,
		Date:                  &blockDate,
	}

	trans := func(ctx context.Context) error {
		afterSwapTvlCount, err := model.UpsertSwapTvlCount(ctx, swapTvlCount)
		if err != nil {
			return errors.Wrap(err)
		}

		_, err = model.UpsertSwapTvlCountDay(ctx, swapTvlCountDay, &blockDate)
		if err != nil {
			return errors.Wrap(err)
		}

		afterUserSwapCount, err := model.UpsertUserSwapCount(ctx, userSwapCount)
		if err != nil {
			return errors.Wrap(err)
		}

		userSwapCountDays, total, err := model.QueryUserSwapCountDay(
			ctx,
			1,
			0,
			model.NewFilter("user_address = ?", userSwapCount.UserAddress),
			model.NewFilter("swap_address = ?", userSwapCount.SwapAddress),
			model.NewFilter("date = ?", blockDate),
		)

		if err != nil {
			return errors.Wrap(err)
		}

		if total == 0 {
			userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume
			userSwapCountDay.MinTxVolume = userSwapCountDay.UserTokenAVolume
		} else {
			if userSwapCountDays[0].MaxTxVolume.LessThan(userSwapCountDay.UserTokenAVolume) {
				userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume
			}

			if userSwapCountDays[0].MinTxVolume.GreaterThan(userSwapCountDay.UserTokenAVolume) {
				userSwapCountDay.MaxTxVolume = userSwapCountDay.UserTokenAVolume
			}
		}

		_, err = model.UpsertUserSwapCountDay(ctx, userSwapCountDay, &blockDate)
		if err != nil {
			return errors.Wrap(err)
		}

		// swap address 最新tvl
		redisKey := domain.SwapTvlCountKey(afterSwapTvlCount.SwapAddress)
		if err = redisClient.Set(ctx, redisKey.Key, afterSwapTvlCount.TokenABalance.String(), redisKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// swap address 总的交易额（vol）
		redisKey = domain.AccountSwapVolCountKey(afterSwapTvlCount.SwapAddress)
		if err = redisClient.Set(ctx, redisKey.Key, afterSwapTvlCount.TokenABalance.String(), redisKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// user address 总的交易额（vol）
		redisKey = domain.AccountSwapVolCountKey(afterUserSwapCount.UserAddress)
		if err = redisClient.Set(ctx, redisKey.Key, afterUserSwapCount.UserTokenAVolume.String(), redisKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		return nil
	}

	if err := model.Transaction(ctx, trans); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// swap 交易量是不能为0的
func getUserSwapVolumeAndBalance(transaction *rpc.TransactionWithMeta) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	if transaction == nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero
	}

	// 是否是swap 交易
	if len(transaction.Meta.PreTokenBalances) != 4 || len(transaction.Meta.PostTokenBalances) != 4 {
		return decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero
	}

	// 更具调用的方法可以确定在 PreTokenBalances / PostTokenBalances 中的第几队表示的是用户的资产信息，或是swap pair 的资产信息
	// 第一二对分表表示的是用户的资产信息。第三四对表示的是swap pair 的资产信息

	var (
		userTokenAPreVolumeStr  string
		userTokenBPreVolumeStr  string
		userTokenAPostVolumeStr string
		userTokenBPostVolumeStr string
		userTokenABalance       decimal.Decimal
		userTokenBalance        decimal.Decimal
	)

	// 此做法只是用于transaction数据是通过GetConfirmedTransaction rpc接口获取的

	userTokenAPreVolumeStr = transaction.Meta.PreTokenBalances[0].UiTokenAmount.Amount
	userTokenBPreVolumeStr = transaction.Meta.PreTokenBalances[1].UiTokenAmount.Amount

	userTokenAPostVolumeStr = transaction.Meta.PostTokenBalances[0].UiTokenAmount.Amount
	userTokenBPostVolumeStr = transaction.Meta.PostTokenBalances[1].UiTokenAmount.Amount

	userTokenAPreVolume, _ := strconv.ParseInt(userTokenAPreVolumeStr, 10, 64)
	userTokenBPreVolume, _ := strconv.ParseInt(userTokenBPreVolumeStr, 10, 64)
	userTokenAPostVolume, _ := strconv.ParseInt(userTokenAPostVolumeStr, 10, 64)
	userTokenBPostVolume, _ := strconv.ParseInt(userTokenBPostVolumeStr, 10, 64)

	userTokenABalance = decimal.NewFromInt(userTokenAPostVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(transaction.Meta.PreTokenBalances[0].UiTokenAmount.Decimals)).Neg()))
	userTokenBalance = decimal.NewFromInt(userTokenBPostVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(transaction.Meta.PreTokenBalances[1].UiTokenAmount.Decimals)).Neg()))

	userTokenADeltaVolume := userTokenAPostVolume - userTokenAPreVolume
	userTokenBDeltaVolume := userTokenBPostVolume - userTokenBPreVolume

	userTokenADeltaVolumeDecimal := decimal.NewFromInt(userTokenADeltaVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(transaction.Meta.PreTokenBalances[0].UiTokenAmount.Decimals)).Neg()))
	userTokenBDeltaVolumeDecimal := decimal.NewFromInt(userTokenBDeltaVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(transaction.Meta.PreTokenBalances[1].UiTokenAmount.Decimals)).Neg()))

	return userTokenADeltaVolumeDecimal, userTokenBDeltaVolumeDecimal, userTokenABalance, userTokenBalance
}
