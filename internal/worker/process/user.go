package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

// UserCount 同步更新swap_counts表和user_swap_counts表
type UserCount struct {
	ID                int64
	LastTransactionID int64
	SwapAccount       string
	//SwapRecords       []*parse.SwapRecord
	tx          *parse.Tx
	BlockDate   *time.Time
	Transaction string
	spec        string
}

func (u *UserCount) getBeginID() error {
	maxID, err := model.GetMaxUserCountKLineID(context.TODO(), u.SwapAccount)
	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}
	u.ID = maxID
	return nil
}

// ParserDate 按照区块时间顺序解析
func (u *UserCount) ParserDate() error {
	for {
		ctx := context.Background()
		if err := u.getBeginID(); err != nil {
			return errors.Wrap(err)
		}

		filters := []model.Filter{
			model.NewFilter("id <= ?", u.LastTransactionID),
			model.SwapAddress(u.SwapAccount),
			model.OrderFilter("id asc"),
			model.NewFilter("id > ?", u.ID),
		}

		swapTransactions, err := model.QuerySwapTransactions(ctx, 100, 0, filters...)
		if err != nil {
			logger.Error("get single transaction err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if len(swapTransactions) == 0 {
			logger.Info(fmt.Sprintf("parse swap, swap address: %u , current id is %d, target id is %d", u.SwapAccount, u.ID, u.LastTransactionID))
			break
		}

		for _, transaction := range swapTransactions {
			u.ID = transaction.ID
			tx := parse.NewTx(transaction.TxData)
			err = tx.ParseTxALl()
			if err != nil {
				if errors.Is(err, errors.RecordNotFound) {
					continue
				}
				logger.Error("sync transaction id err,", logger.Errorf("transaction signature:%s", transaction.Signature), logger.Errorv(err))
				return errors.Wrap(err)
			}

			u.BlockDate = transaction.BlockTime
			u.tx = tx
			u.Transaction = transaction.Signature
			if err = u.WriteToDB(); err != nil {
				return errors.Wrap(err)
			}
		}

		// 更新处理数据的位置
		if err = u.UpdateLastTransActionID(ctx); err != nil {
			return errors.Wrap(err)
		}

		logger.Info(fmt.Sprintf("parse swap, swap address: %u , current id is %d, target id is %d", u.SwapAccount, u.ID, u.LastTransactionID))

	}

	return nil
}

func (u *UserCount) UpdateLastTransActionID(ctx context.Context) error {
	maxID, err := model.GetMaxUserCountKLineID(context.TODO(), u.SwapAccount)
	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if err = model.UpdateUserCountKLine(ctx, map[string]interface{}{"last_swap_transaction_id": u.ID}, model.SwapAddress(u.SwapAccount), model.NewFilter("last_swap_transaction_id = ?", maxID)); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (u *UserCount) WriteToDB() error {
	var (
		err            error
		userCountKLine = &domain.UserCountKLine{}
	)
	trans := func(ctx context.Context) error {
		for _, swapRecord := range u.tx.SwapRecords {
			// 仅当前swapAccount  可以插入
			if swapRecord.SwapConfig.SwapAccount != u.SwapAccount {
				continue
			}

			var (
				tokenAVolume      decimal.Decimal
				tokenBVolume      decimal.Decimal
				tokenAQuoteVolume decimal.Decimal
				tokenBQuoteVolume decimal.Decimal
			)
			if swapRecord.Direction == 0 {
				tokenAVolume = swapRecord.TokenCount.TokenAVolume
				tokenBQuoteVolume = swapRecord.TokenCount.TokenBVolume
			} else {
				tokenBVolume = swapRecord.TokenCount.TokenBVolume
				tokenAQuoteVolume = swapRecord.TokenCount.TokenAVolume
			}

			userCountKLine.UserAddress = swapRecord.UserOwnerAddress
			userCountKLine.SwapAddress = swapRecord.SwapConfig.SwapAccount
			userCountKLine.LastSwapTransactionID = u.ID
			userCountKLine.TokenAAddress = swapRecord.SwapConfig.TokenA.SwapTokenAccount
			userCountKLine.TokenBAddress = swapRecord.SwapConfig.TokenB.SwapTokenAccount
			userCountKLine.Date = u.BlockDate
			userCountKLine.TxNum = 1
			userCountKLine.TokenASymbol = swapRecord.SwapConfig.TokenA.Symbol
			userCountKLine.TokenBSymbol = swapRecord.SwapConfig.TokenB.Symbol

			userCountKLine.UserTokenAVolume = userCountKLine.UserTokenAVolume.Add(tokenAVolume)
			userCountKLine.UserTokenBVolume = userCountKLine.UserTokenBVolume.Add(tokenBVolume)
			userCountKLine.TokenAQuoteVolume = userCountKLine.TokenAQuoteVolume.Add(tokenAQuoteVolume)
			userCountKLine.TokenBQuoteVolume = userCountKLine.TokenBQuoteVolume.Add(tokenBQuoteVolume)
		}

		for _, liquidity := range u.tx.LiquidityRecords {
			// 仅当前swapAccount  可以插入
			if liquidity.SwapConfig.SwapAccount != u.SwapAccount {
				continue
			}

			userCountKLine.UserAddress = liquidity.UserOwnerAddress
			userCountKLine.SwapAddress = liquidity.SwapConfig.SwapAccount
			userCountKLine.LastSwapTransactionID = u.ID
			userCountKLine.Date = u.BlockDate
			userCountKLine.TxNum = 1
			userCountKLine.TokenAAddress = liquidity.SwapConfig.TokenA.SwapTokenAccount
			userCountKLine.TokenBAddress = liquidity.SwapConfig.TokenB.SwapTokenAccount
			userCountKLine.TokenASymbol = liquidity.SwapConfig.TokenA.Symbol
			userCountKLine.TokenBSymbol = liquidity.SwapConfig.TokenB.Symbol
			if liquidity.Direction == 0 {
				userCountKLine.TokenAWithdrawLiquidityVolume = userCountKLine.TokenAWithdrawLiquidityVolume.Add(liquidity.UserCount.TokenAVolume)
				userCountKLine.TokenBWithdrawLiquidityVolume = userCountKLine.TokenBWithdrawLiquidityVolume.Add(liquidity.UserCount.TokenBVolume)

			} else {
				userCountKLine.TokenADepositLiquidityVolume = userCountKLine.TokenADepositLiquidityVolume.Add(liquidity.UserCount.TokenAVolume)
				userCountKLine.TokenBDepositLiquidityVolume = userCountKLine.TokenBDepositLiquidityVolume.Add(liquidity.UserCount.TokenBVolume)
			}
		}

		for _, claim := range u.tx.ClaimRecords {
			// 仅当前swapAccount  可以插入
			if claim.SwapConfig.SwapAccount != u.SwapAccount {
				continue
			}

			userCountKLine.UserAddress = claim.UserOwnerAddress
			userCountKLine.LastSwapTransactionID = u.ID
			userCountKLine.Date = u.BlockDate
			userCountKLine.TxNum = 1
			userCountKLine.SwapAddress = claim.SwapConfig.SwapAccount
			userCountKLine.TokenAAddress = claim.SwapConfig.TokenA.SwapTokenAccount
			userCountKLine.TokenBAddress = claim.SwapConfig.TokenB.SwapTokenAccount
			userCountKLine.TokenASymbol = claim.SwapConfig.TokenA.Symbol
			userCountKLine.TokenBSymbol = claim.SwapConfig.TokenB.Symbol

			userCountKLine.TokenAClaimVolume = userCountKLine.TokenAClaimVolume.Add(claim.UserCount.TokenAVolume)
			userCountKLine.TokenBClaimVolume = userCountKLine.TokenBClaimVolume.Add(claim.UserCount.TokenBVolume)
		}

		newKline := kline.NewKline(u.BlockDate)
		for _, t := range newKline.Types {
			if t.DateType == domain.DateMin || t.DateType == domain.DateTwelfth || t.DateType == domain.DateQuarter || t.DateType == domain.DateHalfAnHour || t.DateType == domain.DateHour {
				continue
			}

			userCountKLine.DateType = t.DateType
			userCountKLine.Date = t.Date
			if err = u.updateUserCountKLine(ctx, userCountKLine, t); err != nil {
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

// updateUserCountKLine 写入user_count_days 表
func (u *UserCount) updateUserCountKLine(ctx context.Context, userCountKline *domain.UserCountKLine, t *kline.Type) error {
	userSwapCountDays, total, err := model.QueryUserSwapCountDay(
		ctx,
		1,
		0,
		model.NewFilter("user_address = ?", userCountKline.UserAddress),
		model.NewFilter("swap_address = ?", userCountKline.SwapAddress),
		model.NewFilter("date = ?", t.Date),
	)

	if err != nil {
		return errors.Wrap(err)
	}

	if total == 0 {
		userCountKline.MaxTxVolume = userCountKline.UserTokenAVolume
		userCountKline.MinTxVolume = userCountKline.UserTokenAVolume
	} else {
		if userSwapCountDays[0].MaxTxVolume.LessThan(userCountKline.UserTokenAVolume) {
			userCountKline.MaxTxVolume = userCountKline.UserTokenAVolume
		}

		if userSwapCountDays[0].MinTxVolume.GreaterThan(userCountKline.UserTokenAVolume) {
			userCountKline.MaxTxVolume = userCountKline.UserTokenAVolume
		}
	}

	_, err = model.UpsertUserSwapCountKLine(ctx, userCountKline)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil

}
