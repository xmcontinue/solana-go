package sol

import (
	"context"
	"fmt"

	"git.cplus.link/go/akit/errors"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/near/borsh-go"
)

var ActivityProgramId = ag_solanago.MustPublicKeyFromBase58("ACTj4dCtGJqDsy7TJ5C3TDbjbeA3nzoaLW9LW571MzeU")


func GetMetadata(mintAddress string) (*Metadata, error) {
	mintKey := ag_solanago.MustPublicKeyFromBase58(mintAddress)
	metaDataAccount, _, err := ag_solanago.FindTokenMetadataAddress(mintKey)
	if err != nil {
		return &Metadata{}, errors.Wrap(err)
	}
	solanaClient := GetRpcClient()
	res, err := solanaClient.GetAccountInfoWithOpts(
		context.TODO(),
		metaDataAccount,
		&rpc.GetAccountInfoOpts{
			Encoding: ag_solanago.EncodingBase64,
		},
	)

	if err != nil {
		return &Metadata{}, errors.Wrap(err)
	}

	metadata, err := MetadataDeserialize(res.Value.Data.GetBinary())
	if err != nil {
		return &Metadata{}, errors.Wrap(err)
	}
	return &metadata, nil
}

func GetActivityMeta(mintAddress string) (*ActivityMetadata, error) {
	mintKey := ag_solanago.MustPublicKeyFromBase58(mintAddress)
	seed := [][]byte{
		[]byte("activity"),
		mintKey[:],
	}
	activityMetaKey, _, err := ag_solanago.FindProgramAddress(seed, ActivityProgramId)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	solanaClient := GetRpcClient()
	res, err := solanaClient.GetAccountInfoWithOpts(
		context.TODO(),
		activityMetaKey,
		&rpc.GetAccountInfoOpts{
			Encoding: ag_solanago.EncodingBase64,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err)
	}

	var metadata ActivityMetadata
	err = borsh.Deserialize(&metadata, res.Value.Data.GetBinary()[8:])
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data, err: %v", err)
	}
	return &metadata, nil
}
