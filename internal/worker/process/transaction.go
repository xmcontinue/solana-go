package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
)

type syncType string

var (
	LastSwapTransactionID syncType = "swap:transaction:last_id"
	// 如果有新增的表，则新增redis key ，用以判断当前表同步数据位置，且LastSwapTransactionID为截止id
)

func transactionIDCache() error {
	// 同步当前进度到redis
	lastSwapTransaction, err := model.QuerySwapTransaction(context.Background(), model.OrderFilter("slot desc,id desc"))
	if err != nil {
		logger.Error("sync transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}
	err = redisClient.Set(context.TODO(), string(LastSwapTransactionID), lastSwapTransaction.ID, 0).Err()
	if err != nil {
		logger.Error("sync transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil
}

func getTransactionID() (int64, error) {
	res := redisClient.Get(context.TODO(), string(LastSwapTransactionID))
	if res.Err() != nil {
		logger.Error("sync transaction id err", logger.Errorv(res.Err()))
		return 0, errors.Wrap(res.Err())
	}

	return res.Int64()
}
