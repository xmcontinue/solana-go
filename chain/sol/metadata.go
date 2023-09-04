package sol

import (
	"context"
	"strings"

	"git.cplus.link/go/akit/errors"
	ag_binary "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func getProgramAccountsResultAccount(dataSize uint64, programIDPub ag_solanago.PublicKey, memCmpAccount string, memCmpAccountOffset uint64) ([]ag_solanago.PublicKey, error) {
	rpcFilters := make([]rpc.RPCFilter, 0, 2)

	if dataSize != 0 {
		rpcFilters = append(rpcFilters, rpc.RPCFilter{
			DataSize: dataSize,
		})
	}

	if memCmpAccount != "" {
		collectionBase58 := ag_solanago.MustPublicKeyFromBase58(memCmpAccount).Bytes()
		rpcFilters = append(rpcFilters, rpc.RPCFilter{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: memCmpAccountOffset,
				Bytes:  collectionBase58[:],
			},
		})
	}

	offset := uint64(0)
	length := uint64(0)
	opt := &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   ag_solanago.EncodingBase64,
		Filters:    rpcFilters,
		DataSlice: &rpc.DataSlice{
			Offset: &offset,
			Length: &length,
		},
	}

	out, err := GetRpcClient().GetProgramAccountsWithOpts(context.Background(), programIDPub, opt)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	accounts := make([]ag_solanago.PublicKey, 0, len(out))
	for _, v := range out {
		accounts = append(accounts, v.Pubkey)
	}

	return accounts, nil
}

func getAccountDataByAccount(resultAccount []ag_solanago.PublicKey) (rpc.GetProgramAccountsResult, error) {
	var finish bool
	var begin = 0
	var limitAccount = 100
	out := make(rpc.GetProgramAccountsResult, 0, len(resultAccount))
	for !finish {
		var accounts []ag_solanago.PublicKey
		if len(resultAccount) < begin+limitAccount {
			finish = true
			accounts = resultAccount[begin:]
		} else {
			accounts = resultAccount[begin : begin+limitAccount]
			begin += limitAccount
		}

		result, err := GetRpcClient().GetMultipleAccountsWithOpts(context.Background(), accounts, nil)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		for i, v := range result.Value {
			out = append(out, &rpc.KeyedAccount{
				Pubkey:  accounts[i],
				Account: v,
			})
		}
	}

	if len(out) == 0 {
		return nil, errors.Wrap(errors.RecordNotFound)
	}

	return out, nil
}

const (
	CremaMetadataCollectionIndex = 402
	CremaMetadataDataSize        = 679
)

func PubMetadataAccount(collectionMintAccount string) (rpc.GetProgramAccountsResult, error) {
	resultAccount, err := getProgramAccountsResultAccount(CremaMetadataDataSize, ag_solanago.TokenMetadataProgramID, collectionMintAccount, CremaMetadataCollectionIndex)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return getAccountDataByAccount(resultAccount)
}

func DecodeMetadata(data []byte) (*token_metadata.Metadata, error) {
	metadata := &token_metadata.Metadata{}
	metadataDecoder := ag_binary.NewDecoderWithEncoding(data, ag_binary.EncodingBorsh)
	err := metadata.UnmarshalWithDecoder(metadataDecoder)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	metadata.Data.Name = strings.Trim(metadata.Data.Name, "\x00")
	metadata.Data.Symbol = strings.Trim(metadata.Data.Symbol, "\x00")
	metadata.Data.Uri = strings.Trim(metadata.Data.Uri, "\x00")

	return metadata, nil
}
