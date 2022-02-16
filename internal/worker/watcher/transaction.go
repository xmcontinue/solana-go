package watcher

import (
	"context"
	"encoding/json"
	"sync"

	"git.cplus.link/go/akit/errors"
	"github.com/gagliardetto/solana-go"

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
	// query the last success sync transaction
	var signature *solana.Signature
	transaction, err := model.QueryBaseTransaction(context.Background())
	if err == nil {
		sig, _ := solana.SignatureFromBase58(transaction.Signature)
		signature = &sig
	}

	// get signature list
	signatures, err := s.tvl.PullSignatures(signature, 10)
	if err != nil {
		return errors.Wrap(err)
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

	return nil
}

func CreateSyncTransaction() error {
	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncTvl", &m, func(value interface{}) JobInterface {
		return &SyncTransaction{
			name: "sync_tvl",
			spec: "0 */10 * * * *",
			tvl:  sol.NewTVL(value.(*sol.SwapConfig)),
		}
	})
	if err != nil {
		return err
	}

	return nil
}
