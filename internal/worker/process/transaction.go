package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func transactionIDCache() error {
	// 同步当前进度到redis
	lastSwapTransaction, err := model.QuerySwapTransaction(context.Background(), model.OrderFilter("id desc"))
	if err != nil {
		logger.Error("sync transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	lastSwapTransactionV2, err := model.GetSwapTransactionV2(context.Background(), model.OrderFilter("id desc"))
	if err != nil {
		if !errors.Is(err, errors.RecordNotFound) {
			logger.Error("sync transaction id err", logger.Errorv(err))
			return errors.Wrap(err)
		}
	} else {
		if lastSwapTransactionV2.ID > lastSwapTransaction.ID {
			lastSwapTransaction.ID = lastSwapTransactionV2.ID
		}
	}

	err = redisClient.Set(context.TODO(), domain.LastSwapTransactionID().Key, lastSwapTransaction.ID, 0).Err()
	if err != nil {
		logger.Error("sync transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil
}

func getTransactionID() (int64, error) {
	res := redisClient.Get(context.TODO(), domain.LastSwapTransactionID().Key)
	if res.Err() != nil {
		logger.Error("sync transaction id err", logger.Errorv(res.Err()))
		return 0, errors.Wrap(res.Err())
	}

	return res.Int64()
}
