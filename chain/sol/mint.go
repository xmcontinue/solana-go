package sol

import (
	"context"

	"git.cplus.link/go/akit/errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetMintsByTokenOwner(ctx context.Context, wallet string) ([]string, error) {
	conf := &rpc.GetTokenAccountsConfig{
		ProgramId: &ag_solanago.TokenProgramID,
	}

	opts := &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   ag_solanago.EncodingBase64,
	}

	out, err := GetRpcClient().GetTokenAccountsByOwner(ctx, ag_solanago.MustPublicKeyFromBase58(wallet), conf, opts)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	mints := make([]string, 0, len(out.Value))
	for _, v := range out.Value {
		accountByte := v.Account.Data.GetBinary()
		account := &token.Account{}
		accountDecoder := ag_binary.NewDecoderWithEncoding(accountByte, ag_binary.EncodingBorsh)
		err = account.UnmarshalWithDecoder(accountDecoder)
		if err != nil {
			return nil, err
		}

		if account.Amount == 1 {
			mints = append(mints, account.Mint.String())
		}
	}

	return mints, nil
}
