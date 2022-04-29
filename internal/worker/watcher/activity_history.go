package watcher

import (
	"context"
	"fmt"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/chain/sol"
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