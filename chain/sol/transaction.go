package sol

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func (tvl *TVL) PullSignatures(before *solana.Signature, limit int) ([]*rpc.TransactionSignature, error) {
	opts := &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentFinalized,
	}

	if before != nil {
		opts.Before = *before
	}

	out, err := tvl.client.GetSignaturesForAddressWithOpts(
		context.TODO(),
		tvl.SwapPublicKey,
		opts,
	)

	if err != nil {
		return nil, errors.Wrap(err)
	}

	return out, nil
}

func (tvl *TVL) GetTransactionsForSignature(signatures []*rpc.TransactionSignature) ([]*rpc.GetTransactionResult, error) {
	transactions := make([]*rpc.GetTransactionResult, 0, len(signatures))

	opts := rpc.GetTransactionOpts{
		Encoding:   solana.EncodingJSON,
		Commitment: rpc.CommitmentMax,
	}

	for _, value := range signatures {
		out, err := tvl.client.GetTransaction(
			context.TODO(),
			value.Signature,
			&opts,
		)

		if err != nil {
			return nil, errors.Wrap(err)
		}

		if err != nil || out.Meta.Err != nil {
			continue
		}

		if len(out.Transaction.GetParsedTransaction().Message.Instructions[0].Data) != 17 {
			continue
		}

		transactions = append(transactions, out)
	}

	return transactions, nil
}
