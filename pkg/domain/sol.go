package domain

import (
	"git.cplus.link/go/akit/util/decimal"
	"github.com/xmcontinue/solana-go"
	"github.com/xmcontinue/solana-go/rpc"
)

type ChainNet struct {
	Address string
	Client  *rpc.Client
	Height  uint64
	Slot    uint64
	Status  uint64
}

type SwapConfig struct {
	Name              string `json:"name" mapstructure:"name"`
	Fee               string `json:"fee" mapstructure:"fee"`
	SwapAccount       string `json:"swap_account" mapstructure:"swap_account"`
	PoolAddress       string `json:"pool_address" mapstructure:"pool_address"`
	SwapPublicKey     solana.PublicKey
	TokenA            Token             `json:"token_a" mapstructure:"token_a"`
	TokenB            Token             `json:"token_b" mapstructure:"token_b"`
	PriceInterval     *PriceInterval    `json:"price_interval" mapstructure:"price_interval"`
	Version           string            `json:"-"    mapstructure:"version"`
	IsPause           bool              `json:"isPause"   mapstructure:"-"`
	RewarderUsd       []decimal.Decimal `json:"-"   mapstructure:"-"`
	ISDisplayRewarder bool              `json:"is_display_rewarder" mapStructure:"is_display_rewarder"`
}

type TokenConfig struct {
	Name   string `json:"name" mapstructure:"name"`
	Symbol string `json:"symbol" mapstructure:"symbol"`
}

type Token struct {
	Symbol                 string `json:"symbol" mapstructure:"symbol"`
	Name                   string `json:"name" mapstructure:"name"`
	TokenMint              string `json:"token_mint" mapstructure:"token_mint"`
	TokenMintPublicKey     solana.PublicKey
	SwapTokenAccount       string `json:"swap_token_account" mapstructure:"swap_token_account"`
	SwapTokenPublicKey     solana.PublicKey
	RefundAddress          string `json:"refund_address" mapstructure:"refund_address"`
	RefundAddressPublicKey solana.PublicKey
	Decimal                uint8           `json:"decimal" mapstructure:"decimal"`
	ShowDecimal            uint8           `json:"show_decimal" mapstructure:"show_decimal"`
	Balance                decimal.Decimal `json:"-"`
	RefundBalance          decimal.Decimal `json:"-"`
}

type TokenInfo struct {
	Symbol  string `json:"symbol" mapstructure:"symbol"`
	Address solana.PublicKey
	Decimal uint8 `json:"decimals" mapstructure:"decimals"`
}

type SwapConfigView struct {
	Name              string         `json:"name" mapstructure:"name"`
	Fee               string         `json:"fee" mapstructure:"fee"`
	SwapAccount       string         `json:"swap_account" mapstructure:"swap_account"`
	TokenA            TokenView      `json:"token_a" mapstructure:"token_a"`
	TokenB            TokenView      `json:"token_b" mapstructure:"token_b"`
	PriceInterval     *PriceInterval `json:"price_interval" mapstructure:"price_interval"`
	IsPause           bool           `json:"isPause"   mapstructure:"-"`
	ISDisplayRewarder bool           `json:"is_display_rewarder" mapStructure:"is_display_rewarder"`
	RewarderDisplay1  bool           `json:"rewarder_display1" mapstructure:"rewarder_display1"`
	RewarderDisplay2  bool           `json:"rewarder_display2" mapstructure:"rewarder_display2"`
	RewarderDisplay3  bool           `json:"rewarder_display3" mapstructure:"rewarder_display3"`
}

type TokenView struct {
	Symbol           string `json:"symbol" mapstructure:"symbol"`
	Name             string `json:"name" mapstructure:"name"`
	TokenMint        string `json:"token_mint" mapstructure:"token_mint"`
	Decimal          uint8  `json:"decimal" mapstructure:"decimal"`
	ShowDecimal      uint8  `json:"show_decimal" mapstructure:"show_decimal"`
	CalculateDecimal uint8  `json:"calculate_decimal" mapstructure:"calculate_decimal"`
}
