package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type parserV1 struct {
	ID                int64
	LastTransactionID int64

	SwapAccount  string
	SwapRecords  []*parse.SwapRecord
	tx           *parse.Tx
	BlockDate    *time.Time
	spec         string
	Version      string
	Transactions []*domain.SwapTransaction
}

func (s *parserV1) GetSwapAccount() string {
	return s.SwapAccount
}

func (s *parserV1) GetBeginId() (int64, error) {
	return s.Transactions[len(s.Transactions)-1].ID, nil
}

func (s *parserV1) GetTransactions(limit, offset int, filters ...model.Filter) error {
	swapTransactions, err := model.QuerySwapTransactions(context.TODO(), limit, offset, filters...)
	if err != nil {
		logger.Error("get single transaction err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	if len(swapTransactions) == 0 {
		logger.Info(fmt.Sprintf("parse swap, swap address: %s , current id is %d, target id is %d", s.SwapAccount, s.ID, s.LastTransactionID))
		return errors.RecordNotFound
	}

	s.Transactions = swapTransactions
	return nil
}

func (s *parserV1) GetParsingCutoffID() int64 {
	return s.LastTransactionID
}

func (s *parserV1) UpdateLastTransActionID() error {
	// 更新处理数据的位置
	if err := model.UpdateSwapCountBySwapAccount(context.TODO(), s.SwapAccount, map[string]interface{}{"last_swap_transaction_id": s.ID}); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *parserV1) ParserAllInstructionType() error {
	for _, transaction := range s.Transactions {
		s.ID = transaction.ID

		tx := parse.NewTx(transaction.TxData)
		err := tx.ParseTxALl()
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			logger.Error("sync transaction id err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		writeTyp := &WriteTyp{
			ID:          s.ID,
			SwapAccount: s.SwapAccount,
			BlockDate:   transaction.BlockTime,
		}

		var isContinue = true
		if len(tx.SwapRecords) != 0 {
			swapRecordIface := make([]parse.SwapRecordIface, len(tx.SwapRecords), len(tx.SwapRecords))
			for i, v := range tx.SwapRecords {
				swapRecordIface[i] = v
			}
			writeTyp.swapRecords = swapRecordIface
			isContinue = false
		}

		if len(tx.LiquidityRecords) != 0 {
			liquidityIface := make([]parse.LiquidityRecordIface, len(tx.LiquidityRecords), len(tx.LiquidityRecords))
			for i, v := range tx.LiquidityRecords {
				liquidityIface[i] = v
			}
			writeTyp.liquidityRecords = liquidityIface
			isContinue = false
		}

		if len(tx.ClaimRecords) != 0 {
			claimRecords := make([]parse.CollectRecordIface, len(tx.ClaimRecords), len(tx.ClaimRecords))
			for i, v := range tx.ClaimRecords {
				claimRecords[i] = v
			}
			writeTyp.claimRecords = claimRecords
			isContinue = false
		}

		if isContinue {
			continue
		}

		if err := WriteAllToDB(writeTyp); err != nil {
			logger.Error("write to db error:", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}
	return nil
}

func (s *parserV1) ParserSwapInstruction() error {
	for _, transaction := range s.Transactions {
		s.ID = transaction.ID

		tx := parse.NewTx(transaction.TxData)
		err := tx.ParseTxToSwap()
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			logger.Error("sync transaction id err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		swapRecordIface := make([]parse.SwapRecordIface, len(tx.SwapRecords), len(tx.SwapRecords))
		for i, v := range tx.SwapRecords {
			swapRecordIface[i] = v
		}

		writeTyp := &WriteTyp{
			ID:          s.ID,
			SwapAccount: s.SwapAccount,
			BlockDate:   transaction.BlockTime,
			swapRecords: swapRecordIface,
		}
		if err := WriteSwapRecordToDB(writeTyp, transaction.TokenAUSD, transaction.TokenBUSD); err != nil {
			logger.Error("write to db error:", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}

	return nil
}
