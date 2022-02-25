package watcher

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// SyncTransaction sync transaction
type SyncTransaction struct {
	name string
	spec string
	tvl  *sol.TVL
}

func (s *SyncTransaction) Name() string {
	return s.name
}

func (s *SyncTransaction) GetSpecFunc() string {
	return s.spec
}

func (s *SyncTransaction) DeleteJobFunc(_ *JobInfo) error {
	return nil
}

func CreateSyncTransaction() error {
	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncTransaction", &m, func(value interface{}) JobInterface {
		return &SyncTransaction{
			name: "sync_transaction",
			spec: getSpec("sync_transaction"),
			tvl:  sol.NewTVL(value.(*sol.SwapConfig)),
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *SyncTransaction) Run() error {
	complete := false
	for {
		err := s.SyncTransaction(&complete)
		if err != nil {
			break
		}
		if complete {
			break
		}
	}

	return nil
}

func (s *SyncTransaction) SyncTransaction(complete *bool) error {
	// get signatures for swap account address
	before, until, err := s.getBeforeAndUntil()
	if err != nil {
		return errors.Wrap(err)
	}

	signatures, err := s.getSignatures(before, until, complete)
	if err != nil {
		return errors.Wrap(err)
	}
	if *complete == true {
		return nil
	}

	// get transactions for signatures
	transactions, err := s.tvl.GetTransactionsForSignature(signatures)
	if err != nil {
		return errors.Wrap(err)
	}

	// update data to pgsql
	err = s.writeTxToDb(before, until, signatures, transactions)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// getBeforeAndUntil
func (s *SyncTransaction) getBeforeAndUntil() (*solana.Signature, *solana.Signature, error) {
	// create before, until
	var (
		before *solana.Signature
		until  *solana.Signature
	)
	swapPairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddress(s.tvl.SwapAccount))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = model.CreateSwapPairBase(context.Background(), &domain.SwapPairBase{
				SwapAddress:   s.tvl.SwapAccount,
				TokenAAddress: s.tvl.TokenA.SwapTokenAccount,
				TokenBAddress: s.tvl.TokenB.SwapTokenAccount,
				IsSync:        false,
			})
			if err != nil {
				return nil, nil, errors.Wrap(err)
			}
		} else {
			return nil, nil, errors.Wrap(err)
		}
	} else {
		if swapPairBase.IsSync {
			sig, _ := solana.SignatureFromBase58(swapPairBase.EndSignature)
			until = &sig
		} else {
			sig, _ := solana.SignatureFromBase58(swapPairBase.StartSignature)
			before = &sig
		}
	}

	return before, until, nil
}

// getSignatures
func (s *SyncTransaction) getSignatures(before *solana.Signature, until *solana.Signature, complete *bool) ([]*rpc.TransactionSignature, error) {
	// get signature list (max limit is 1000 )
	limit := 1000
	signatures, err := s.tvl.PullSignatures(before, until, limit)
	if err != nil {
		return signatures, errors.Wrap(err)
	}

	if len(signatures) == 0 {
		// sync finished
		if before != nil {
			err = model.UpdateSwapPairBase(
				context.Background(),
				map[string]interface{}{
					"is_sync": true,
				},
				model.SwapAddress(s.tvl.SwapAccount),
			)
			if err != nil {
				return signatures, errors.Wrap(err)
			}
		}

		*complete = true

		return signatures, nil
	}

	isComplete := len(signatures) < limit
	if before == nil && until != nil && !isComplete {
		// sync back
		afterSignatures, afterBefore := make([]*rpc.TransactionSignature, 0), signatures[len(signatures)-1].Signature

		for !isComplete {

			newSignatures, err := s.tvl.PullSignatures(&afterBefore, until, limit)
			if err != nil {
				return signatures, errors.Wrap(err)
			}

			isComplete = len(newSignatures) < limit

			if !isComplete {
				afterBefore = newSignatures[len(newSignatures)-1].Signature
			}

			afterSignatures = append(afterSignatures, newSignatures...)
		}

		signatures = append(signatures, afterSignatures...)

	}

	// array inversion
	for i := 0; i < len(signatures)/2; i++ {
		signatures[len(signatures)-1-i], signatures[i] = signatures[i], signatures[len(signatures)-1-i]
	}

	if len(signatures) > 5000 {
		signatures = signatures[:5000]
	}

	return signatures, nil
}

// writeTxToDb
func (s *SyncTransaction) writeTxToDb(before *solana.Signature, until *solana.Signature, signatures []*rpc.TransactionSignature, transactions []*rpc.GetTransactionResult) error {
	// open model transaction
	txModelTransaction := func(mCtx context.Context) error {
		// update schedule
		swapPairBaseMap := map[string]interface{}{}
		if before == nil {
			swapPairBaseMap["end_signature"] = signatures[len(signatures)-1].Signature.String()
		}

		if until == nil {
			swapPairBaseMap["start_signature"] = signatures[0].Signature.String()
		}

		failedNum := len(signatures) - len(transactions)
		if failedNum > 0 {
			swapPairBaseMap["failed_tx_num"] = gorm.Expr("failed_tx_num + ?", failedNum)
		}

		err := model.UpdateSwapPairBase(mCtx, swapPairBaseMap, model.SwapAddress(s.tvl.SwapAccount))
		if err != nil {
			return errors.Wrap(err)
		}

		if len(transactions) == 0 {
			return nil
		}

		// created transaction record
		swapTransactions := make([]*domain.SwapTransaction, 0, len(transactions))

		for _, v := range transactions {
			blockTime := v.BlockTime.Time()

			data := domain.TxData(*v)

			tokenAVolume, tokenBVolume, tokenABalance, tokenBBalance := s.getSwapVolume(v)

			swapTransactions = append(swapTransactions, &domain.SwapTransaction{
				Signature:      v.Transaction.GetParsedTransaction().Signatures[0].String(),
				Fee:            precisionConversion(decimal.NewFromInt(int64(v.Meta.Fee)), 9),
				BlockTime:      &blockTime,
				Slot:           v.Slot,
				UserAddress:    "",
				InstructionLen: getInstructionLen(v.Transaction.GetParsedTransaction().Message.Instructions),
				SwapAddress:    s.tvl.SwapAccount,
				TokenAAddress:  s.tvl.TokenA.SwapTokenAccount,
				TokenBAddress:  s.tvl.TokenB.SwapTokenAccount,
				TokenAVolume:   tokenAVolume,
				TokenBVolume:   tokenBVolume,
				TokenABalance:  tokenABalance,
				TokenBBalance:  tokenBBalance,
				Status:         true,
				TxData:         &data,
			})
		}

		transactionsLen := len(swapTransactions)

		page := 100
		for i := 0; i < transactionsLen; i = i + page {
			if transactionsLen < i+page {
				err = model.CreateSwapTransactions(mCtx, swapTransactions[i:transactionsLen])
				if err != nil {
					return errors.Wrap(err)
				}
				break
			} else {
				err = model.CreateSwapTransactions(mCtx, swapTransactions[i:i+page])
				if err != nil {
					return errors.Wrap(err)
				}
			}
		}

		return nil
	}

	err := model.Transaction(context.Background(), txModelTransaction)
	if err != nil {
		time.Sleep(time.Second * 5)
		return errors.Wrap(err)
	}

	logger.Info(fmt.Sprintf("sync transaction : swap account(%s) signature from %s to %s", s.tvl.SwapAccount, signatures[0].Signature.String(), signatures[len(signatures)-1].Signature.String()))

	return nil
}

func (s *SyncTransaction) getSwapVolume(meta *rpc.GetTransactionResult) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	var (
		tokenAPreBalanceStr  string
		tokenBPreBalanceStr  string
		tokenAPostBalanceStr string
		tokenBPostBalanceStr string
	)

	for _, tokenBalance := range meta.Meta.PreTokenBalances {
		keyIndex := tokenBalance.AccountIndex
		key := meta.Transaction.GetParsedTransaction().Message.AccountKeys[keyIndex]
		if key.Equals(s.tvl.TokenA.SwapTokenPublicKey) {
			tokenAPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(s.tvl.TokenB.SwapTokenPublicKey) {
			tokenBPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	for _, tokenBalance := range meta.Meta.PostTokenBalances {
		keyIndex := tokenBalance.AccountIndex
		key := meta.Transaction.GetParsedTransaction().Message.AccountKeys[keyIndex]
		if key.Equals(s.tvl.TokenA.SwapTokenPublicKey) {
			tokenAPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(s.tvl.TokenB.SwapTokenPublicKey) {
			tokenBPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	tokenAPreBalance, _ := decimal.NewFromString(tokenAPreBalanceStr)
	tokenAPostBalance, _ := decimal.NewFromString(tokenAPostBalanceStr)
	tokenBPreBalance, _ := decimal.NewFromString(tokenBPreBalanceStr)
	tokenBPostBalance, _ := decimal.NewFromString(tokenBPostBalanceStr)

	tokenADeltaVolume, tokenBDeltaVolume := tokenAPostBalance.Sub(tokenAPreBalance), tokenBPostBalance.Sub(tokenBPreBalance)

	return precisionConversion(tokenADeltaVolume, int(s.tvl.TokenA.Decimal)),
		precisionConversion(tokenBDeltaVolume, int(s.tvl.TokenB.Decimal)),
		precisionConversion(tokenAPostBalance, int(s.tvl.TokenA.Decimal)),
		precisionConversion(tokenBPostBalance, int(s.tvl.TokenB.Decimal))
}

// precisionConversion 精度转换
func precisionConversion(num decimal.Decimal, precision int) decimal.Decimal {
	return num.Div(decimal.NewFromFloat(math.Pow10(precision)))
}

// getInstructionLen 获取第一个Instruction data长度
func getInstructionLen(instructions []rpc.ParsedInstruction) uint64 {
	for _, v := range instructions {
		dataLen := len(v.Data)
		if dataLen > 0 {
			return uint64(dataLen)
		}
	}
	return 0
}
