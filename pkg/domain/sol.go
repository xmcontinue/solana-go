package domain

import (
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type ChainNet struct {
	Address string
	Client  *rpc.Client
	Height  uint64
	Slot    uint64
	Status  uint64
}

type SwapConfig struct {
	Name          string `json:"name" mapstructure:"name"`
	Fee           string `json:"fee" mapstructure:"fee"`
	SwapAccount   string `json:"swap_account" mapstructure:"swap_account"`
	PoolAddress   string `json:"pool_address" mapstructure:"pool_address"`
	SwapPublicKey solana.PublicKey
	TokenA        Token          `json:"token_a" mapstructure:"token_a"`
	TokenB        Token          `json:"token_b" mapstructure:"token_b"`
	PriceInterval *PriceInterval `json:"price_interval" mapstructure:"price_interval"`
}

type Token struct {
	Symbol             string `json:"symbol" mapstructure:"symbol"`
	TokenMint          string `json:"token_mint" mapstructure:"token_mint"`
	TokenMintPublicKey solana.PublicKey
	SwapTokenAccount   string `json:"swap_token_account" mapstructure:"swap_token_account"`
	SwapTokenPublicKey solana.PublicKey
	Decimal            uint8           `json:"decimal" mapstructure:"decimal"`
	Balance            decimal.Decimal `json:"-"`
}
