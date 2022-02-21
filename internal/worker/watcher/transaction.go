package watcher

import (
	"context"
	"encoding/json"
	"sync"

	"git.cplus.link/go/akit/errors"
	"github.com/gagliardetto/solana-go"
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

func (s *SyncTransaction) Run() error {
	// create before, until
	before, until := &solana.Signature{}, &solana.Signature{}
	swapPairBase, err := model.QuerySwapPairBase(context.Background(), model.SwapAddress(s.tvl.SwapAccount))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.CreateSwapPairBase(context.Background(), &domain.SwapPairBase{
			SwapAddress:   s.tvl.SwapAccount,
			TokenAAddress: s.tvl.TokenA.SwapTokenAccount,
			TokenBAddress: s.tvl.TokenB.SwapTokenAccount,
			IsSync:        false,
		})
		if err != nil {
			return errors.Wrap(err)
		}
	} else {
		if swapPairBase.IsSync {
			*before, _ = solana.SignatureFromBase58(swapPairBase.EndSignature)
		} else {
			*until, _ = solana.SignatureFromBase58(swapPairBase.StartSignature)
		}
	}

	// get signature list
	signatures, err := s.tvl.PullSignatures(before, until, 10)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(signatures) == 0 {
		// sync finished
		if before == nil {
			return nil
		} else {
			err = model.UpdateSwapPairBase(
				context.Background(),
				map[string]interface{}{
					"is_sync": true,
				},
				model.SwapAddress(s.tvl.SwapAccount),
			)
			if err != nil {
				return errors.Wrap(err)
			}

			return nil
		}
	}

	// array inversion
	for i := 0; i < len(signatures)/2; i++ {
		signatures[len(signatures)-1-i], signatures[i] = signatures[i], signatures[len(signatures)-1-i]
	}

	// get transaction list
	transactions, err := s.tvl.GetTransactionsForSignature(signatures)
	if err != nil {
		return errors.Wrap(err)
	}

	baseTransactions := make([]*domain.TransactionBase, 0, len(transactions))

	for _, v := range transactions {
		blockTime := v.BlockTime.Time()
		transactionData, _ := v.Transaction.MarshalJSON()
		metaData, _ := json.Marshal(v.Meta)

		baseTransactions = append(baseTransactions, &domain.TransactionBase{
			BlockTime:       &blockTime,
			Slot:            v.Slot,
			TransactionData: string(transactionData),
			MateData:        string(metaData),
			Signature:       v.Transaction.GetParsedTransaction().Signatures[0].String(),
		})
	}
	if len(baseTransactions) == 0 {
		return nil
	}

	// created transaction record
	err = model.CreateBaseTransactions(context.Background(), baseTransactions)
	if err != nil {
		return errors.Wrap(err)
	}

	// update schedule
	swapPairBaseMap := map[string]interface{}{
		"start_signature": signatures[len(signatures)-1].Signature.String(),
		"end_signature":   signatures[0].Signature.String(),
	}

	err = model.UpdateSwapPairBase(
		context.Background(),
		swapPairBaseMap,
		model.SwapAddress(s.tvl.SwapAccount),
	)
	if err != nil {
		return errors.Wrap(err)
	}

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
