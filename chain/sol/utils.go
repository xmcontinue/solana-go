package sol

import (
	"fmt"
	"strings"

	"github.com/near/borsh-go"
	ag_solanago "github.com/xmcontinue/solana-go"
)

type Key borsh.Enum

type Creator struct {
	Address  ag_solanago.PublicKey
	Verified bool
	Share    uint8
}

type Data struct {
	Name                 string
	Symbol               string
	Uri                  string
	SellerFeeBasisPoints uint16
	Creators             *[]Creator
}

type Metadata struct {
	Key                 Key
	UpdateAuthority     ag_solanago.PublicKey
	Mint                ag_solanago.PublicKey
	Data                Data
	PrimarySaleHappened bool
	IsMutable           bool
	EditionNonce        *uint8
}

func MetadataDeserialize(data []byte) (Metadata, error) {
	var metadata Metadata
	err := borsh.Deserialize(&metadata, data)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to deserialize data, err: %v", err)
	}
	// trim null byte
	metadata.Data.Name = strings.TrimRight(metadata.Data.Name, "\x00")
	metadata.Data.Symbol = strings.TrimRight(metadata.Data.Symbol, "\x00")
	metadata.Data.Uri = strings.TrimRight(metadata.Data.Uri, "\x00")
	return metadata, nil
}

type ActivityMetadata struct {
	Caffeine             uint64
	Degree               uint8
	Mint                 ag_solanago.PublicKey
	MintUser             ag_solanago.PublicKey
	Seed                 uint8
	IsCrmClaimed         bool
	IsSecondPartyClaimed bool
}
