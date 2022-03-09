package sol

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func PullSignatures(key solana.PublicKey, before *solana.Signature, until *solana.Signature, limit int) ([]*rpc.TransactionSignature, error) {
	opts := &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentFinalized,
	}

	if before != nil {
		opts.Before = *before
	}

	if until != nil {
		opts.Until = *until
	}

	out, err := GetRpcClient().GetSignaturesForAddressWithOpts(
		context.TODO(),
		key,
		opts,
	)

	if err != nil {
		return nil, errors.Wrap(err)
	}

	return out, nil
}

func GetTransactionsForSignature(signatures []*rpc.TransactionSignature) ([]*rpc.GetTransactionResult, error) {
	transactions := make([]*rpc.GetTransactionResult, 0, len(signatures))

	opts := rpc.GetTransactionOpts{
		Encoding:   solana.EncodingJSON,
		Commitment: rpc.CommitmentMax,
	}

	for _, value := range signatures {
		if value.Err != nil {
			continue
		}
		out, err := GetRpcClient().GetTransaction(
			context.TODO(),
			value.Signature,
			&opts,
		)

		if err != nil {
			return nil, errors.Wrap(err)
		}

		transactions = append(transactions, out)
	}

	return transactions, nil
}
