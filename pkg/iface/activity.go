package iface

import (
	"time"

	ag_solanago "github.com/gagliardetto/solana-go"
)

type GetActivityNftMetadataReq struct {
	Mint string `json:"mint"`
}

type Creator struct {
	Address ag_solanago.PublicKey `json:"address"`
	Share   uint8                 `json:"share"`
}

type File struct {
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type Properties struct {
	Category string     `json:"category"`
	Creators *[]Creator `json:"creators"`
	Files    *[]File    `json:"files"`
}

type GetActivityNftMetadataResp struct {
	Name                 string      `json:"name"`
	Symbol               string      `json:"symbol"`
	Image                string      `json:"image"`
	SellerFeeBasisPoints uint16      `json:"seller_fee_basis_points"`
	Properties           *Properties `json:"properties"`
}

type CreateActivityHistoryReq struct {
}

type CreateActivityHistoryResp struct {
}

type GetActivityHistoryByUserReq struct {
	User string `json:"user"`
}

type ActivityHistoryResp struct {
	ID           int64 `json:"id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	UserKey      string     `json:"user_key"`
	MintKey      string     `json:"mint_key"`
	Crm          float64    `json:"crm"`
	Marinade     float64    `json:"marinade"`
	Port         float64    `json:"port"`
	Hubble       float64    `json:"hubble"`
	Nirv         float64    `json:"nirv"`
	SignatureCrm string     `json:"signature_crm"`
	Signature    string     `json:"signature"`
	BlockTime    int64      `json:"blocktime"`
	Degree       uint8      `json:"degree"`
	Caffeine     uint64     `json:"caffeine"`
}

type GetActivityHistoryByUserResp struct {
	History []*ActivityHistoryResp `json:"history"`
}
