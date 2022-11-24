package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// SyncTransaction sync transaction
type SyncTransaction struct {
	name       string
	spec       string
	swapConfig *domain.SwapConfig
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
			name:       "sync_transaction",
			spec:       getSpec("sync_transaction"),
			swapConfig: value.(*domain.SwapConfig),
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
			logger.Error(fmt.Sprintf("sync failed: swap account(%s) ", s.swapConfig.SwapAccount), logger.Errorv(err))
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
	transactions, err := sol.GetTransactionsForSignature(signatures)
	if err != nil {
		return errors.Wrap(err)
	}

	if s.swapConfig.Version == "v2" {
		err = s.writeTxToDbV2(before, until, signatures, transactions)
	} else {
		err = s.writeTxToDb(before, until, signatures, transactions)
	}
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
	swapPairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddressFilter(s.swapConfig.SwapAccount))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = model.CreateSwapPairBase(context.Background(), &domain.SwapPairBase{
				SwapAddress:      s.swapConfig.SwapAccount,
				TokenAAddress:    s.swapConfig.TokenA.SwapTokenAccount,
				TokenBAddress:    s.swapConfig.TokenB.SwapTokenAccount,
				IsSync:           false,
				SyncUtilFinished: true,
			})
			if err != nil {
				return nil, nil, errors.Wrap(err)
			}
		} else {
			return nil, nil, errors.Wrap(err)
		}
	} else {
		if swapPairBase.IsSync {
			if swapPairBase.EndSignature != "" {
				sig, _ := solana.SignatureFromBase58(swapPairBase.EndSignature)
				until = &sig
			}
		} else {
			if swapPairBase.StartSignature != "" {
				sig, _ := solana.SignatureFromBase58(swapPairBase.StartSignature)
				before = &sig
			}
		}
	}

	return before, until, nil
}

// getSignatures
func (s *SyncTransaction) getSignatures(before *solana.Signature, until *solana.Signature, complete *bool) ([]*rpc.TransactionSignature, error) {
	// get signature list (max limit is 1000 )
	limit := 1000
	signatures, err := sol.PullSignatures(s.swapConfig.SwapPublicKey, before, until, limit)
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
				model.SwapAddressFilter(s.swapConfig.SwapAccount),
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
			logger.Info(fmt.Sprintf("sync signatures: swap account(%s) ,from(%s),to(%s) ", s.swapConfig.SwapAccount, afterBefore.String(), until))

			newSignatures, err := sol.PullSignatures(s.swapConfig.SwapPublicKey, &afterBefore, until, limit)
			if err != nil {
				return signatures, errors.Wrap(err)
			}

			if len(newSignatures) == 0 {
				break
			}

			isComplete = len(newSignatures) < limit

			if !isComplete {
				afterBefore = newSignatures[len(newSignatures)-1].Signature
			}

			afterSignatures = append(afterSignatures, newSignatures...)

			if len(afterSignatures) > 15000 {
				afterSignatures = afterSignatures[5000:]
			}
		}

		// signatures = append(signatures, afterSignatures...)
		//signatures = afterSignatures
		if len(afterSignatures) != 0 {
			signatures = afterSignatures
		}
	}

	// array inversion
	for i, j := 0, len(signatures)-1; i < j; i, j = i+1, j-1 {
		signatures[i], signatures[j] = signatures[j], signatures[i]
	}

	// if len(signatures) > limit {
	// 	signatures = signatures[:limit]
	// }

	return signatures, nil
}

func (s *SyncTransaction) parserTx(transactions []*rpc.GetTransactionResult) []*domain.SwapTransactionV2 {
	if len(transactions) == 0 {
		return nil
	}

	swapTransactionV2 := make([]*domain.SwapTransactionV2, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.Meta.Err != nil {
			continue
		}

		tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(transaction.Transaction.GetBinary()))
		if err != nil {
			logger.Error("TransactionFromDecoder error", logger.Errorv(err))
			continue
		}

		blockTime := transaction.BlockTime.Time()

		logs, _ := json.Marshal(transaction.Meta.LogMessages)
		txV2 := parse.NewTxV2()
		err = txV2.ParseAllV2(string(logs))
		if err != nil {
			continue
		}

		txType, userAccount := getTxTypeAndUserAccountV2(txV2)

		swapTransactionV2 = append(swapTransactionV2, &domain.SwapTransactionV2{
			SwapAddress: s.swapConfig.SwapAccount,
			UserAddress: userAccount,
			Signature:   tx.Signatures[0].String(),
			FeePayer:    tx.Message.AccountKeys[0].String(),
			Slot:        transaction.Slot,
			BlockTime:   &blockTime,
			Msg:         string(logs),
			TxType:      txType,
		})
	}

	return swapTransactionV2
}

func getTxTypeAndUserAccountV2(tx *parse.Txv2) (string, string) {
	txTypeMap := make(map[string]struct{})
	var userAccount string
	if len(tx.SwapRecords) != 0 {
		for _, v := range tx.SwapRecords {
			txTypeMap[v.EventName] = struct{}{}
			userAccount = v.GetUserAddress()
		}
	}

	if len(tx.LiquidityRecords) != 0 {
		for _, v := range tx.LiquidityRecords {
			txTypeMap[v.EventName] = struct{}{}
			userAccount = v.GetUserAddress()
		}
	}

	if len(tx.ClaimRecords) != 0 {
		for _, v := range tx.SwapRecords {
			txTypeMap[v.EventName] = struct{}{}
			userAccount = v.GetUserAddress()
		}
	}

	var tType string
	for v := range txTypeMap {
		tType += v + ","
	}

	tType = strings.TrimRight(tType, ",")
	return tType, userAccount
}
func (s *SyncTransaction) writeTxToDbV2(before *solana.Signature, until *solana.Signature, signatures []*rpc.TransactionSignature, transactions []*rpc.GetTransactionResult) error {
	swapTransactionV2s := s.parserTx(transactions)

	tokenAUSD, err := model.GetPriceForSymbol(context.Background(), s.swapConfig.TokenA.Symbol)
	if err != nil {
		return errors.Wrap(err)
	}

	tokenBUSD, err := model.GetPriceForSymbol(context.Background(), s.swapConfig.TokenB.Symbol)
	if err != nil {
		return errors.Wrap(err)
	}

	trans := func(ctx context.Context) error {
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

		err = model.UpdateSwapPairBase(ctx, swapPairBaseMap, model.SwapAddressFilter(s.swapConfig.SwapAccount))
		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapTransactionV2s) == 0 {
			return nil
		}

		for _, v := range swapTransactionV2s {
			_, err = model.GetSwapTransactionV2(ctx, model.SwapAddressFilter(v.SwapAddress), model.NewFilter("signature = ?", v.Signature))
			if !errors.Is(err, errors.RecordNotFound) {
				continue
			}

			tx := parse.NewTxV2()
			err = tx.ParseAllV2(v.Msg)
			if err == nil {
				v.TxType, v.UserAddress = getTxTypeAndUserAccountV2(tx)
			}

			v.TokenAUSD = tokenAUSD
			v.TokenBUSD = tokenBUSD

			err = model.CreatedSwapTransactionV2(ctx, v)
			if err != nil {
				return errors.Wrap(err)
			}
		}

		return nil
	}

	err = model.Transaction(context.Background(), trans)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// writeTxToDb
func (s *SyncTransaction) writeTxToDb(before *solana.Signature, until *solana.Signature, signatures []*rpc.TransactionSignature, transactions []*rpc.GetTransactionResult) error {
	if len(signatures) == 0 {
		return errors.Wrap(errors.New(fmt.Sprintf("signatures is zero, swap_account: %s", s.swapConfig.SwapAccount)))
	}

	tokenAUSD, err := model.GetPriceForSymbol(context.Background(), s.swapConfig.TokenA.Symbol)
	if err != nil {
		return errors.Wrap(err)
	}

	tokenBUSD, err := model.GetPriceForSymbol(context.Background(), s.swapConfig.TokenB.Symbol)
	if err != nil {
		return errors.Wrap(err)
	}
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

		err = model.UpdateSwapPairBase(mCtx, swapPairBaseMap, model.SwapAddressFilter(s.swapConfig.SwapAccount))
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

			var txType, userAccount string
			tx := parse.NewTx(&data)
			err = tx.ParseTxALl()
			if err == nil {
				txType, userAccount = getTxTypeAndUserAccount(tx)
			}

			tokenAVolume, tokenBVolume, tokenABalance, tokenBBalance := s.getSwapVolume(v, tx.TransAction)

			swapTransactions = append(swapTransactions, &domain.SwapTransaction{
				Signature:      tx.TransAction.Signatures[0].String(),
				Fee:            parse.PrecisionConversion(decimal.NewFromInt(int64(v.Meta.Fee)), 9),
				BlockTime:      &blockTime,
				Slot:           v.Slot,
				UserAddress:    userAccount,
				InstructionLen: getCompiledInstructionLen(tx.TransAction.Message.Instructions),
				SwapAddress:    s.swapConfig.SwapAccount,
				TokenAAddress:  s.swapConfig.TokenA.SwapTokenAccount,
				TokenBAddress:  s.swapConfig.TokenB.SwapTokenAccount,
				TokenAVolume:   tokenAVolume,
				TokenBVolume:   tokenBVolume,
				TokenABalance:  tokenABalance,
				TokenBBalance:  tokenBBalance,
				TokenAUSD:      tokenAUSD,
				TokenBUSD:      tokenBUSD,
				Status:         true,
				TxData:         &data,
				TxType:         txType,
			})
		}

		// transactionsLen := len(swapTransactions)

		for _, v := range swapTransactions {
			_, err = model.QuerySwapTransaction(context.Background(), model.SwapAddressFilter(v.SwapAddress), model.NewFilter("signature = ?", v.Signature))
			if err != nil {
				err = model.CreateSwapTransactions(mCtx, []*domain.SwapTransaction{v})
				if err != nil {
					return errors.Wrap(err)
				}
			}
		}

		// page := 100
		// for i := 0; i < transactionsLen; i = i + page {
		// 	if transactionsLen < i+page {
		// 		err = model.CreateSwapTransactions(mCtx, swapTransactions[i:transactionsLen])
		// 		if err != nil {
		// 			return errors.Wrap(err)
		// 		}
		// 		break
		// 	} else {
		// 		err = model.CreateSwapTransactions(mCtx, swapTransactions[i:i+page])
		// 		if err != nil {
		// 			return errors.Wrap(err)
		// 		}
		// 	}
		// }

		return nil
	}

	err = model.Transaction(context.Background(), txModelTransaction)
	if err != nil {
		time.Sleep(time.Second * 5)
		return errors.Wrap(err)
	}

	logger.Info(fmt.Sprintf("sync transaction : swap account(%s) signature from %s to %s", s.swapConfig.SwapAccount, signatures[0].Signature.String(), signatures[len(signatures)-1].Signature.String()))

	return nil
}

func getTxTypeAndUserAccount(tx *parse.Tx) (string, string) {
	txTypeMap := make(map[string]struct{})
	var userAccount string
	if len(tx.SwapRecords) != 0 {
		txTypeMap[parse.SwapType] = struct{}{}
		userAccount = tx.SwapRecords[0].GetUserAddress()
	}

	if len(tx.LiquidityRecords) != 0 {
		for _, v := range tx.LiquidityRecords {
			if v.Direction == 0 {
				txTypeMap[parse.DecreaseLiquidityType] = struct{}{}
			} else {
				txTypeMap[parse.IncreaseLiquidityType] = struct{}{}
			}
			userAccount = v.GetUserAddress()
		}
	}

	if len(tx.ClaimRecords) != 0 {
		txTypeMap[parse.ClaimType] = struct{}{}
		userAccount = tx.ClaimRecords[0].GetUserAddress()
	}

	var tType string
	for v := range txTypeMap {
		tType += v + ","
	}

	tType = strings.TrimRight(tType, ",")
	return tType, userAccount
}

func (s *SyncTransaction) getSwapVolume(meta *rpc.GetTransactionResult, tx *solana.Transaction) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	var (
		tokenAPreBalanceStr  string
		tokenBPreBalanceStr  string
		tokenAPostBalanceStr string
		tokenBPostBalanceStr string
	)

	for _, tokenBalance := range meta.Meta.PreTokenBalances {
		keyIndex := tokenBalance.AccountIndex

		key := tx.Message.AccountKeys[keyIndex]
		if key.Equals(s.swapConfig.TokenA.SwapTokenPublicKey) {
			tokenAPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(s.swapConfig.TokenB.SwapTokenPublicKey) {
			tokenBPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	for _, tokenBalance := range meta.Meta.PostTokenBalances {
		keyIndex := tokenBalance.AccountIndex

		key := tx.Message.AccountKeys[keyIndex]
		if key.Equals(s.swapConfig.TokenA.SwapTokenPublicKey) {
			tokenAPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(s.swapConfig.TokenB.SwapTokenPublicKey) {
			tokenBPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	tokenAPreBalance, _ := decimal.NewFromString(tokenAPreBalanceStr)
	tokenAPostBalance, _ := decimal.NewFromString(tokenAPostBalanceStr)
	tokenBPreBalance, _ := decimal.NewFromString(tokenBPreBalanceStr)
	tokenBPostBalance, _ := decimal.NewFromString(tokenBPostBalanceStr)

	tokenADeltaVolume, tokenBDeltaVolume := tokenAPostBalance.Sub(tokenAPreBalance), tokenBPostBalance.Sub(tokenBPreBalance)

	return parse.PrecisionConversion(tokenADeltaVolume, int(s.swapConfig.TokenA.Decimal)),
		parse.PrecisionConversion(tokenBDeltaVolume, int(s.swapConfig.TokenB.Decimal)),
		parse.PrecisionConversion(tokenAPostBalance, int(s.swapConfig.TokenA.Decimal)),
		parse.PrecisionConversion(tokenBPostBalance, int(s.swapConfig.TokenB.Decimal))
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

func getCompiledInstructionLen(instructions []solana.CompiledInstruction) uint64 {
	for _, v := range instructions {
		dataLen := len(v.Data)
		if dataLen > 0 {
			return uint64(dataLen)
		}
	}
	return 0
}
