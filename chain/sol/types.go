package sol

import (
	"encoding/json"

	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
)

type MetadataJSON struct {
	Name                 string       `json:"name"`
	Symbol               string       `json:"symbol"`
	Description          string       `json:"description"`
	SellerFeeBasisPoints uint16       `json:"seller_fee_basis_points"`
	Image                string       `json:"image"`
	Attributes           *[]Attribute `json:"attributes"`
	Properties           *Properties  `json:"properties"`
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

type Properties struct {
	Creators *[]Creators `json:"creators"`
	Files    *[]Files    `json:"files"`
}

type Creators struct {
	Address string `json:"address"`
	Share   uint8  `json:"share"`
}

type Files struct {
	URI  string `json:"uri"`
	Type string `json:"type"`
}

type Gallery struct {
	*token_metadata.Metadata `json:"*_token___metadata_._metadata,omitempty"`
	*MetadataJSON            `json:"*_metadata_json,omitempty"`
	Name                     string `json:"name,omitempty" json:"name,omitempty"`
	Owner                    string `json:"owner,omitempty"`
	Mint                     string `json:"mint,omitempty"`
}

type MainGallery struct {
	Name  string `json:"name,omitempty" json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
	Mint  string `json:"mint,omitempty"`
}

func (s Gallery) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s Gallery) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s MainGallery) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s MainGallery) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
