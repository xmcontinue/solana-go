package sol

import (
	"context"

	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetOwnerByMintAccount(mintAccount ag_solanago.PublicKey) (out *rpc.GetTokenLargestAccountsResult, err error) {
	return GetRpcClient().GetTokenLargestAccounts(context.Background(), mintAccount, rpc.CommitmentConfirmed)
}

func GetAccountInfoWithOpts(account ag_solanago.PublicKey) (out *rpc.GetAccountInfoResult, err error) {
	return GetRpcClient().GetAccountInfoWithOpts(context.Background(), account, &rpc.GetAccountInfoOpts{
		Encoding:   ag_solanago.EncodingBase64,
		Commitment: rpc.CommitmentConfirmed,
	})
}
