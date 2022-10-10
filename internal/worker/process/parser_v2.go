package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type parserV2 struct {
	ID                int64
	LastTransactionID int64
	SwapAccount       string
	Version           string
	TransactionV2s    []*domain.SwapTransactionV2
	Transactions      []*domain.SwapTransaction
	BlockDate         *time.Time
	//SwapRecords       []*parse.SwapRecordV2
	swapR []parse.SwapRecordIface
}

func (s *parserV2) GetParsingCutoffID() int64 {
	return s.LastTransactionID
}

func (s *parserV2) GetSwapAccount() string {
	return s.SwapAccount
}

func (s *parserV2) GetTransactions(limit, offset int, filters ...model.Filter) error {
	swapTransactions, err := model.QuerySwapTransactionsV2(context.TODO(), limit, offset, filters...)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(swapTransactions) == 0 {
		return errors.RecordNotFound
	}

	s.TransactionV2s = swapTransactions
	return nil
}

func getBeginID(swapAccount string) (int64, error) {
	swapCount, err := model.QuerySwapCount(context.TODO(), model.SwapAddressFilter(swapAccount))
	if err != nil {
		if !errors.Is(err, errors.RecordNotFound) {
			return 0, errors.Wrap(err)
		}
		return 0, nil
	}

	var maxId int64
	if swapCount != nil {
		maxId = swapCount.LastSwapTransactionID
	}
	return maxId, nil
}

func (s *parserV2) GetBeginId() (int64, error) {
	return s.TransactionV2s[len(s.TransactionV2s)-1].ID, nil
}

func (s *parserV2) UpdateLastTransActionID() error {
	// 更新处理数据的位置
	if err := model.UpdateSwapCountBySwapAccount(context.TODO(), s.SwapAccount, map[string]interface{}{"last_swap_transaction_id": s.ID}); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *parserV2) ParserAllInstructionType() error {
	for _, transaction := range s.TransactionV2s {
		s.ID = transaction.ID
		tx := parse.NewTxV2()
		err := tx.ParseAllV2(transaction.Msg)
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

		if err := writeAllToDB(writeTyp); err != nil {
			logger.Error("write to db error:", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}
	return nil
}

// ParserSwapInstruction 解析swap
func (s *parserV2) ParserSwapInstruction() error {
	for _, transaction := range s.TransactionV2s {
		s.ID = transaction.ID

		tx := parse.NewTxV2()
		if transaction.Signature != "2ztE9ZM9TB1oX1bbkBcbacfyzo2GHHp2Hgn7mwTzt5CT1SScdrEozoWLeBWvjTFRZsN11xvgp3BSDskmyYgnh1g8" {
			continue
		}
		err := tx.ParseSwapV2(transaction.Msg)
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			logger.Error("sync transaction id err", logger.Errorv(err))
			return errors.Wrap(err)
		}

		if len(tx.SwapRecords) == 0 {
			continue
		}

		swapRecordIface := make([]parse.SwapRecordIface, len(tx.SwapRecords), len(tx.SwapRecords))
		for i := range tx.SwapRecords {
			swapRecordIface[i] = tx.SwapRecords[i]
		}

		writeTyp := &WriteTyp{
			ID:                    transaction.ID,
			LastSwapTransactionID: s.LastTransactionID,
			SwapAccount:           s.SwapAccount,
			BlockDate:             transaction.BlockTime,
			swapRecords:           swapRecordIface,
		}
		if err := writeSwapRecordToDB(writeTyp, transaction.TokenAUSD, transaction.TokenBUSD); err != nil {
			logger.Error("write to db error:", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}

	return nil
}
