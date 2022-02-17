package domain

import (
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type SwapTransaction struct {
	ID            int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt     *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt     *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	Signature     string          `json:"signature" gorm:"not null;type:varchar(64);  index"`       // 交易签名
	Fee           decimal.Decimal `json:"fee" gorm:"type:decimal(36,18)"`                           // 手续费
	BlockTime     *time.Time      `json:"block_time" gorm:"not null;type:timestamp(6)"`             // 打包时间
	Slot          uint64          `json:"slot"  gorm:"not null"`                                    // 区块高度
	UserAddress   string          `json:"user_address" gorm:"not null;type:varchar(64);  index"`    // 用户账户
	SwapAddress   string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume  decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume  decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	Status        bool            `json:"status"`                                                   // 交易装填: 0-失败，1-成功
	TxData        string          `json:"tx_data"       gorm:"not null;type:varchar(512)" `         // 原数据（json格式）
}

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
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	Tvl                   decimal.Decimal `json:"tvl"  gorm:"type:decimal(36,18);default:0"`                // tvl
	Vol                   decimal.Decimal `json:"vol"  gorm:"type:decimal(36,18);default:0"`                // vol
}

type SwapTvlCountDay struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	Tvl                   decimal.Decimal `json:"tvl"  gorm:"type:decimal(36,18);default:0"`                // tvl
	Vol                   decimal.Decimal `json:"vol"  gorm:"type:decimal(36,18);default:0"`                // vol
	Date                  *time.Time      `json:"date" gorm:"not null;type:timestamp(6);index"`             // 统计日期
	TxNum                 int64           `json:"tx_num"`                                                   // 交易笔数
}

type UserSwapCount struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	UserAddress           string          `json:"user_address" gorm:"not null;type:varchar(64);  index"`    // 用户 address
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	TxNum                 int64           `json:"tx_num"`                                                   // 交易笔数
	MaxTxVolume           decimal.Decimal `json:"max_tx_volume" gorm:"type:decimal(36,18);default:0"`       // 最大交易额
	MinTxVolume           decimal.Decimal `json:"min_tx_volume"  gorm:"type:decimal(36,18);default:0"`      // 最小交易额
}

type UserSwapCountDay struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	UserAddress           string          `json:"user_address" gorm:"not null;type:varchar(64);  index"`    // 用户 address
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	TxNum                 int64           `json:"tx_num"`                                                   // 交易笔数
	MaxTxVolume           decimal.Decimal `json:"max_tx_volume" gorm:"type:decimal(36,18);default:0"`       // 最大交易额
	MinTxVolume           decimal.Decimal `json:"min_tx_volume"  gorm:"type:decimal(36,18);default:0"`      // 最小交易额
	Date                  *time.Time      `json:"date" gorm:"not null;type:timestamp(6);index"`             // 统计日期
}
