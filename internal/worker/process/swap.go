package process

import (
	"context"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type SwapTx struct {
	Transaction *domain.SwapTransaction // 可以使用里面的swap address tokenAAddress tokenBAddress
	TokenAIndex int64
	TokenBIndex int64
	BlockDate   *time.Time
}

func (s *SwapTx) Parser() error {
	swapTvlCount := s.NewSwapTransactionTvlCount()
	swapTvlCountDay := s.NewSwapTransactionTvlCountDay()
	userSwapCount, userSwapCountDay := s.NewUserSwapCountAndDay()

	trans := func(ctx context.Context) error {
		afterSwapTvlCount, err := model.UpsertSwapTvlCount(ctx, swapTvlCount)
		if err != nil {
			return errors.Wrap(err)
		}

		_, err = model.UpsertSwapTvlCountDay(ctx, swapTvlCountDay, s.BlockDate)
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
			model.NewFilter("date = ?", s.BlockDate),
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

		_, err = model.UpsertUserSwapCountDay(ctx, userSwapCountDay, s.BlockDate)
		if err != nil {
			return errors.Wrap(err)
		}

		// swap address 最新tvl,单位是价格
		swapTvlKey := domain.SwapTvlCountKey(afterSwapTvlCount.SwapAddress)
		if err = redisClient.Set(ctx, swapTvlKey.Key, afterSwapTvlCount.Tvl.String(), swapTvlKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// swap address 总的交易额（vol），单位是价格
		swapVolKey := domain.AccountSwapVolCountKey(afterSwapTvlCount.SwapAddress)

		if err = redisClient.Set(ctx, swapVolKey.Key, afterSwapTvlCount.Vol.String(), swapVolKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// user address 总的交易额（vol）
		userVolKey := domain.AccountSwapVolCountKey(afterUserSwapCount.UserAddress)
		if err = redisClient.Set(ctx, userVolKey.Key, afterUserSwapCount.UserTokenAVolume.Add(afterUserSwapCount.UserTokenBVolume).String(), userVolKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		return nil
	}

	if err := model.Transaction(context.TODO(), trans); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// NewUserSwapCountAndDay 解析出不同情况下用户的交易额和余额
func (s *SwapTx) NewUserSwapCountAndDay() (*domain.UserSwapCount, *domain.UserSwapCountDay) {
	var (
		userTokenAPreVolumeStr  string
		userTokenBPreVolumeStr  string
		userTokenAPostVolumeStr string
		userTokenBPostVolumeStr string
		userTokenABalance       decimal.Decimal
		userTokenBBalance       decimal.Decimal
	)

	for _, v := range s.Transaction.TxData.Meta.PreTokenBalances {
		if s.TokenAIndex == int64(v.AccountIndex) {
			userTokenAPreVolumeStr = v.UiTokenAmount.Amount
			continue
		}

		if s.TokenBIndex == int64(v.AccountIndex) {
			userTokenBPreVolumeStr = v.UiTokenAmount.Amount
			continue
		}

	}

	for _, v := range s.Transaction.TxData.Meta.PostTokenBalances {
		if s.TokenAIndex == int64(v.AccountIndex) {
			userTokenAPostVolumeStr = v.UiTokenAmount.Amount
			continue
		}

		if s.TokenBIndex == int64(v.AccountIndex) {
			userTokenBPostVolumeStr = v.UiTokenAmount.Amount
			continue
		}

	}

	userTokenAPreVolume, _ := strconv.ParseInt(userTokenAPreVolumeStr, 10, 64)
	userTokenBPreVolume, _ := strconv.ParseInt(userTokenBPreVolumeStr, 10, 64)
	userTokenAPostVolume, _ := strconv.ParseInt(userTokenAPostVolumeStr, 10, 64)
	userTokenBPostVolume, _ := strconv.ParseInt(userTokenBPostVolumeStr, 10, 64)

	userTokenABalance = decimal.NewFromInt(userTokenAPostVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(6))))
	userTokenBBalance = decimal.NewFromInt(userTokenBPostVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(6))))

	userTokenADeltaVolume := userTokenAPostVolume - userTokenAPreVolume
	userTokenBDeltaVolume := userTokenBPostVolume - userTokenBPreVolume

	userTokenADeltaVolumeDecimal := decimal.NewFromInt(userTokenADeltaVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(6))))
	userTokenBDeltaVolumeDecimal := decimal.NewFromInt(userTokenBDeltaVolume).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(6))))

	userSwapCount := &domain.UserSwapCount{
		LastSwapTransactionID: s.Transaction.ID,
		UserAddress:           s.Transaction.UserAddress,
		SwapAddress:           s.Transaction.SwapAddress,
		TokenAAddress:         s.Transaction.TokenAAddress,
		TokenBAddress:         s.Transaction.TokenBAddress,
		UserTokenAVolume:      userTokenADeltaVolumeDecimal,
		UserTokenBVolume:      userTokenBDeltaVolumeDecimal,
		UserTokenABalance:     userTokenABalance,
		UserTokenBBalance:     userTokenBBalance,
		TxNum:                 1,
	}

	// 统计用户每日swap count
	userSwapCountDay := &domain.UserSwapCountDay{
		LastSwapTransactionID: s.Transaction.ID,
		UserAddress:           s.Transaction.UserAddress,
		SwapAddress:           s.Transaction.SwapAddress,
		TokenAAddress:         s.Transaction.TokenAAddress,
		TokenBAddress:         s.Transaction.TokenBAddress,
		UserTokenAVolume:      userTokenADeltaVolumeDecimal,
		UserTokenBVolume:      userTokenBDeltaVolumeDecimal,
		UserTokenABalance:     userTokenABalance,
		UserTokenBBalance:     userTokenBBalance,
		TxNum:                 1,
		Date:                  s.BlockDate,
	}

	return userSwapCount, userSwapCountDay
}

func (s *SwapTx) NewSwapTransactionTvlCount() *domain.SwapTvlCount {
	swapTvlCount := &domain.SwapTvlCount{
		LastSwapTransactionID: s.Transaction.ID,
		SwapAddress:           s.Transaction.SwapAddress,
		TokenAAddress:         s.Transaction.TokenAAddress,
		TokenBAddress:         s.Transaction.TokenBAddress,
		TokenAVolume:          s.Transaction.TokenAVolume,
		TokenBVolume:          s.Transaction.TokenBVolume,
		TokenABalance:         s.Transaction.TokenABalance,
		TokenBBalance:         s.Transaction.TokenBBalance,
		Tvl:                   s.Transaction.TokenABalance.Mul(s.Transaction.TokenAUSD).Add(s.Transaction.TokenABalance.Mul(s.Transaction.TokenBUSD)),
		Vol:                   s.Transaction.TokenAVolume.Mul(s.Transaction.TokenAUSD).Add(s.Transaction.TokenBVolume.Mul(s.Transaction.TokenBUSD)),
	}

	return swapTvlCount
}

func (s *SwapTx) NewSwapTransactionTvlCountDay() *domain.SwapTvlCountDay {

	return &domain.SwapTvlCountDay{
		LastSwapTransactionID: s.Transaction.ID,
		SwapAddress:           s.Transaction.SwapAddress,
		TokenAAddress:         s.Transaction.TokenAAddress,
		TokenBAddress:         s.Transaction.TokenBAddress,
		TokenAVolume:          s.Transaction.TokenAVolume,
		TokenBVolume:          s.Transaction.TokenBVolume,
		TokenABalance:         s.Transaction.TokenABalance,
		TokenBBalance:         s.Transaction.TokenBBalance,
		Tvl:                   s.Transaction.TokenABalance.Mul(s.Transaction.TokenAUSD).Add(s.Transaction.TokenABalance.Mul(s.Transaction.TokenBUSD)),
		Vol:                   s.Transaction.TokenAVolume.Mul(s.Transaction.TokenAUSD).Add(s.Transaction.TokenBVolume.Mul(s.Transaction.TokenBUSD)),
		Date:                  s.BlockDate,
		TxNum:                 1,
	}
}
