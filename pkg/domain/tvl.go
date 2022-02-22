package domain

import (
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type NetRecode struct {
	ID               int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt        *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt        *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	RpcURL           string     `json:"rpc_url"  gorm:"not null;type:varchar(128)"` // Rpc地址
	LastSlot         int64      `json:"last_slot" gorm:"not null"`                  // 最新区块高度
	FailedRequestNum int64      `json:"failed_request_num" gorm:"not null"`         // 请求失败次数
	Enable           bool       `json:"enable" gorm:"default:false"`                // 是否已启用，默认不启用
}

type SwapTvlCount struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`                                                 // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:swap_tvl_count_swap_address_unique_key"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                           // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                           // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                               // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                               // swap token b 余额
	Tvl                   decimal.Decimal `json:"tvl"  gorm:"type:decimal(36,18);default:0"`                                                          // tvl
	Vol                   decimal.Decimal `json:"vol"  gorm:"type:decimal(36,18);default:0"`                                                          // vol
}

type SwapTvlCountDay struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`                                                          // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:swap_tvl_count_day_date_swap_address_unique_key"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                                    // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                                    // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                         // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                         // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                                        // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                                        // swap token b 余额
	Tvl                   decimal.Decimal `json:"tvl"  gorm:"type:decimal(36,18);default:0"`                                                                   // tvl
	Vol                   decimal.Decimal `json:"vol"  gorm:"type:decimal(36,18);default:0"`                                                                   // vol
	Date                  *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:swap_tvl_count_day_date_swap_address_unique_key"`          // 统计日期
	TxNum                 int64           `json:"tx_num"`                                                                                                      // 交易笔数
}

type Tvl struct {
	ID            int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt     *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt     *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	TotalTvlInUsd string     `json:"total_tvl_in_usd" gorm:"type:varchar(32);"`
	TotalVolInUsd string     `json:"total_vol_in_usd" gorm:"type:varchar(32);"`
	Pairs         JsonString `json:"pairs" gorm:"type:text;"`
}

type PairTvl struct {
	Name        string `json:"name"`
	TvlInUsd    string `json:"tvl_in_usd"`
	VolInUsd    string `json:"vol_in_usd"`
	TxNum       uint64 `json:"tx_num"`
	Apr         string `json:"apr"`
	SwapAccount string `json:"swap_account"`
}
