package process

import (
	"context"
	"fmt"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"

	model "git.cplus.link/crema/backend/internal/model/market"
)

type syncType string

var (
	LastSwapTransactionID syncType = "last:swap:transaction:id"
	// 如果有新增的表，则新增redis key ，用以判断当前表同步数据位置，且LastSwapTransactionID为截止id
)

type SwapTokenIndex struct {
	TokenAIndex int64
	TokenBIndex int64
}

var (
	OriginalSwap = &SwapTokenIndex{3, 4}

	JupeterSwap = &SwapTokenIndex{3, 4}
)

// ParserTransaction 不同的transaction使用同一套接口，再解析出transaction后做选择
type ParserTransaction interface {
	Parser(tx *rpc.GetTransactionResult) error
}

var swapInstructionLenMap = map[int]*SwapTokenIndex{
	8:  OriginalSwap,
	26: JupeterSwap,
	17: JupeterSwap,
	12: JupeterSwap,
}

func syncData() error {
	var lastSwapTransactionID int64
	swapTvlCount, err := model.GetLastSwapTvlCount(context.TODO())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("get last transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	if swapTvlCount != nil {
		lastSwapTransactionID = swapTvlCount.LastSwapTransactionID
	}

	for {
		swapTransactions, _, err := model.QuerySwapTransactions(context.TODO(), 1, 0, model.NewFilter("id > ?", lastSwapTransactionID))
		if err != nil {
			logger.Error("get single transaction err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if len(swapTransactions) == 0 {
			break
		}

		for _, transaction := range swapTransactions {

			// parser instructions
			for _, instruction := range transaction.TxData.Transaction.GetParsedTransaction().Message.Instructions {

				tokenIndex, ok := swapInstructionLenMap[len(instruction.Data)]

				if ok {

					if tokenIndex.TokenAIndex == 3 && tokenIndex.TokenBIndex == 4 {
						fmt.Println(tokenIndex.TokenAIndex)
					}

					swapTx := &SwapTx{
						Transaction: transaction,
						TokenAIndex: tokenIndex.TokenAIndex,
						TokenBIndex: tokenIndex.TokenBIndex,
					}

					if err = swapTx.Parser(); err != nil {
						logger.Error("parser data err", logger.Errorv(err))
						return errors.Wrap(err)
					}

					continue
				}

				// 其他方法

			}

			// 同步当前处理数据进度到redis
			lastSwapTransactionID = swapTransactions[len(swapTransactions)-1].ID
			err = redisClient.Set(context.TODO(), string(LastSwapTransactionID), lastSwapTransactionID, 0).Err()
			if err != nil {
				logger.Error("sync transaction id err", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}

	}

	return nil
}
