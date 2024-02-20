package sol

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"github.com/xmcontinue/solana-go"
	"github.com/xmcontinue/solana-go/rpc"
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
	MaxSupportedTransactionVersion := uint64(0)
	opts := rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase58,
		Commitment:                     rpc.CommitmentFinalized,
		MaxSupportedTransactionVersion: &MaxSupportedTransactionVersion,
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
