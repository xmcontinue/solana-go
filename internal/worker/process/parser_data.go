package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"

	model "git.cplus.link/crema/backend/internal/model/market"
)

type syncType string

type SwapTokenIndex struct {
	TokenAIndex int64
	TokenBIndex int64
}

var (
	cremaSwap = &SwapTokenIndex{3, 4}
)

// ParserTransaction 不同的transaction使用同一套接口，再解析出transaction后做选择
type ParserTransaction interface {
	Parser(tx *rpc.GetTransactionResult) error
}

func parserData() error {
	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}
	swapTvlCount, err := model.GetLastSwapTvlCount(context.TODO())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("get last transaction id err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	var beforeSwapTransactionID int64
	if swapTvlCount != nil {
		beforeSwapTransactionID = swapTvlCount.LastSwapTransactionID
	}

	for {
		swapTransactions, err := model.QuerySwapTransactions(context.TODO(), 10, 0, model.NewFilter("id > ?", beforeSwapTransactionID), model.NewFilter("id <= ?", lastSwapTransactionID))
		if err != nil {
			logger.Error("get single transaction err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if len(swapTransactions) == 0 {
			break
		}

		for _, transaction := range swapTransactions {
			blockDate := time.Date(transaction.BlockTime.Year(), transaction.BlockTime.Month(), transaction.BlockTime.Day(), 0, 0, 0, 0, time.UTC)
			// parser instructions
			accountKeys := transaction.TxData.Transaction.GetParsedTransaction().Message.AccountKeys
			for _, instruction := range transaction.TxData.Transaction.GetParsedTransaction().Message.Instructions {
				// 仅已知的swap address 才可以解析
				fmt.Println(accountKeys[instruction.ProgramIDIndex].String())
				if _, ok := swapAccountMap[accountKeys[instruction.ProgramIDIndex].String()]; !ok {
					continue
				}

				// swap 函数在合约里面的下标是1
				if instruction.Data[0] != 1 {
					continue
				}

				swapTx := &SwapTx{
					Transaction: transaction,
					TokenAIndex: instruction.Accounts[cremaSwap.TokenAIndex],
					TokenBIndex: instruction.Accounts[cremaSwap.TokenBIndex],
					BlockDate:   &blockDate,
				}

				if err = swapTx.Parser(); err != nil {
					logger.Error("parser data err", logger.Errorv(err))
					return errors.Wrap(err)
				}
			}

			for _, innerInstruction := range transaction.TxData.Meta.InnerInstructions {
				// 仅已知的swap address 才可以解析
				for _, compiledInstruction := range innerInstruction.Instructions {
					if _, ok := swapAccountMap[accountKeys[compiledInstruction.ProgramIDIndex].String()]; !ok {
						continue
					}

					// swap 函数在合约里面的下标是1
					if compiledInstruction.Data[0] != 1 {
						continue
					}

					swapTx := &SwapTx{
						Transaction: transaction,
						TokenAIndex: int64(compiledInstruction.Accounts[cremaSwap.TokenAIndex]),
						TokenBIndex: int64(compiledInstruction.Accounts[cremaSwap.TokenBIndex]),
						BlockDate:   &blockDate,
					}

					if err = swapTx.Parser(); err != nil {
						logger.Error("parser data err", logger.Errorv(err))
						return errors.Wrap(err)
					}
				}
			}

		}

	}

	return nil
}
