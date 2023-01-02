package sol

import (
	"encoding/json"
	"strings"

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
	Metadata     *token_metadata.Metadata `json:"metadata,omitempty"`
	MetadataJSON *MetadataJSON            `json:"metadata_json,omitempty"`
	Name         string                   `json:"name,omitempty"`
	Mint         string                   `json:"mint,omitempty"`
	Owner        string                   `json:"owner"`
}

// 按照 Person.Age 从大到小排序
type GallerySlice []Gallery

func (a GallerySlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a GallerySlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a GallerySlice) Less(i, j int) bool {
	return strings.Compare(a[j].Name, a[i].Name) > 0 // 重写 Less() 方法， 从大到小排序

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
