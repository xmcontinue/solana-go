package watcher

import (
	"context"
	"fmt"
	"sync"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func SyncActivityTransaction() error {
	// get signatures for swap account address
	until, err := getUntil()
	if err != nil {
		return errors.Wrap(err)
	}
	var before *solana.Signature
	for {
		if before != nil {
			fmt.Println(before.String())
		}
		afterbefore, signatures, err := getSignatures(before, until)
		if err != nil {
			return errors.Wrap(err)
		}
		before = afterbefore
		fmt.Println(signatures[0].Signature.String())
		fmt.Println(signatures[len(signatures)-1].Signature.String())
		// get transactions for signatures
		transactions, err := sol.GetTransactionsForSignature(signatures)
		if err != nil {
			return errors.Wrap(err)
		}

		// update data to pgsql
		err = writeEventToDb(signatures, transactions)

		if err != nil {
			return errors.Wrap(err)
		}

		if before == nil {
			return nil
		}
	}
}

func getUntil() (*solana.Signature, error) {
	var (
		until *solana.Signature
	)
	history, err := model.SelectLatest(context.Background())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if history == nil {
		return nil, nil
	}
	var siga solana.Signature
	if history.SignatureCrm == "" {
		siga, _ = solana.SignatureFromBase58(history.Signature)
	} else {
		siga, _ = solana.SignatureFromBase58(history.SignatureCrm)
	}
	until = &siga
	return until, nil
}

// getSignatures
func getSignatures(before *solana.Signature, until *solana.Signature) (*solana.Signature, []*rpc.TransactionSignature, error) {
	// get signature list (max limit is 1000 )
	limit := 100
	signatures, err := sol.PullSignatures(sol.ActivityProgramId, before, until, limit)
	if err != nil {
		return nil, signatures, errors.Wrap(err)
	}

	if len(signatures) < limit {
		return nil, signatures, nil
	}
	afterBefore := signatures[len(signatures)-1].Signature
	return &afterBefore, signatures, nil
}

func writeEventToDb(signatures []*rpc.TransactionSignature, transactions []*rpc.GetTransactionResult) error {
	for _, e := range transactions {
		blockTime := e.BlockTime.Time()
		events, err := sol.GetAcEventParser().Decode(e.Meta.LogMessages)
		if err != nil {
			return errors.Wrap(err)
		}
		if len(events) == 0 {
			continue
		}
		event := events[0]
		history, err := model.SelectByUserMint(context.Background(), event.User, event.Mint)
		if err != nil {
			continue
		}
		if history == nil {
			if event.EventName == "ClaimRewardEvent" {
				model.CreateActivityHistory(context.Background(), &domain.ActivityHistory{
					UserKey:      event.User,
					MintKey:      event.Mint,
					Crm:          event.Amount,
					SignatureCrm: e.Transaction.GetParsedTransaction().Signatures[0].String(),
					BlockTime:    blockTime.Unix(),
					Degree:       event.Degree,
					Caffeine:     event.Caffeine,
				})
			}
			if event.EventName == "ClaimSecondPartyEvent" {
				model.CreateActivityHistory(context.Background(), &domain.ActivityHistory{
					UserKey:   event.User,
					MintKey:   event.Mint,
					Marinade:  event.Amounts[0],
					Port:      event.Amounts[1],
					Hubble:    event.Amounts[2],
					Nirv:      event.Amounts[3],
					Signature: e.Transaction.GetParsedTransaction().Signatures[0].String(),
					BlockTime: blockTime.Unix(),
					Degree:    event.Degree,
					Caffeine:  event.Caffeine,
				})
			}
		} else {
			if event.EventName == "ClaimRewardEvent" {
				model.UpdateActivityHistory(context.Background(), history.ID, map[string]interface{}{
					"Crm":          event.Amount,
					"SignatureCrm": e.Transaction.GetParsedTransaction().Signatures[0].String(),
				})
			}
			if event.EventName == "ClaimSecondPartyEvent" {
				model.UpdateActivityHistory(context.Background(), history.ID, map[string]interface{}{
					"Marinade":  event.Amounts[0],
					"Port":      event.Amounts[1],
					"Hubble":    event.Amounts[2],
					"Nirv":      event.Amounts[3],
					"Signature": e.Transaction.GetParsedTransaction().Signatures[0].String(),
				})
			}
		}

	}

	logger.Info(fmt.Sprintf("sync activity signature from %s to %s", signatures[0].Signature.String(), signatures[len(signatures)-1].Signature.String()))

	return nil
}

func SyncTypeAndUserAddressHistory() error {
	ctx := context.Background()
	swapPairs, err := model.QuerySwapPairBases(ctx, 1000, 0, model.NewFilter("sync_util_finished = false"))
	if err != nil {
		return errors.Wrap(err)
	}

	if len(swapPairs) == 0 {
		return nil
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(swapPairs))
	for i, swapPair := range swapPairs {

		if swapPair.SyncUtilID == 0 {
			filters := []model.Filter{
				model.SwapAddressFilter(swapPair.SwapAddress),
				model.NewFilter("tx_type != ?", ""),
				model.OrderFilter("id asc"),
			}
			swapTransaction, err := model.QuerySwapTransaction(ctx, filters...)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("QuerySwapTransaction err", logger.Errorv(err))
				return errors.Wrap(err)
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				swapTransaction, err = model.QuerySwapTransaction(ctx,
					model.SwapAddressFilter(swapPair.SwapAddress),
					model.OrderFilter("id desc"),
				)
			}

			err = model.UpdateSwapPairBase(ctx, map[string]interface{}{
				"sync_util_id": swapTransaction.ID,
			},
				model.SwapAddressFilter(swapPair.SwapAddress),
			)
			if err != nil {
				return errors.Wrap(err)
			}

			swapPairs[i].SyncUtilID = swapTransaction.ID
		}

		go func() {
			err = SyncTypeAndUserAddressSingle(swapPairs[i], wg)
			if err != nil {
				logger.Error("SyncTypeAndUserAddressSingle err", logger.Errorv(err))
				return
			}
		}()
	}
	wg.Wait()
	logger.Warn("携程结束。。。。。。。。。。。。。。。。")
	return nil
}

func SyncTypeAndUserAddressSingle(swapPair *domain.SwapPairBase, wg *sync.WaitGroup) error {
	defer func() {
		fmt.Println("++++++++++")
		wg.Done()
	}()
	fmt.Println("1111111111")
	ctx := context.Background()
	beginID := int64(0)
	logger.Warn("开始。。。", logger.String("", swapPair.SwapAddress))
	for {
		filters := []model.Filter{
			model.SwapAddressFilter(swapPair.SwapAddress),
			model.NewFilter("id > ?", beginID),
			model.NewFilter("user_address =''"),
			model.NewFilter("id < ?", swapPair.SyncUtilID),
			model.OrderFilter("id asc"),
		}

		swapTransactions, err := model.QuerySwapTransactions(context.Background(), 1000, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapTransactions) == 0 {
			break
		}
		logger.Warn("开始。。。2", logger.String("", swapPair.SwapAddress))
		for _, swapTransaction := range swapTransactions {
			tx := parse.NewTx(swapTransaction.TxData)
			logger.Warn("开始。。。3", logger.String("", swapPair.SwapAddress))
			err = tx.ParseTxALl()
			if err != nil {
				continue
			}

			txType, userAccount := getTxTypeAndUserAccount(tx)
			err = model.UpdateSwapTransaction(ctx, map[string]interface{}{
				"user_address": userAccount,
				"tx_type":      txType,
			},
				model.IDFilter(swapTransaction.ID),
			)
			logger.Warn("开始。。。4", logger.String(userAccount, swapPair.SwapAddress))
			if err != nil {
				return errors.Wrap(err)
			}
			logger.Warn("开始。。。5", logger.String(userAccount, swapPair.SwapAddress))
		}

		beginID = swapTransactions[len(swapTransactions)-1].ID
	}

	err := model.UpdateSwapPairBase(ctx, map[string]interface{}{
		"sync_util_finished": true,
	},
		model.SwapAddressFilter(swapPair.SwapAddress),
	)

	return errors.Wrap(err)
}
