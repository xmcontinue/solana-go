package domain

import (
	"database/sql/driver"
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type SwapPairCount struct {
	ID                int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt         *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt         *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	TokenAVolume      decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);"`
	TokenBVolume      decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);"`
	TokenABalance     decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);"`
	TokenBBalance     decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);"`
	TokenAPoolAddress string          `json:"-" gorm:"type:varchar(64);  index"`
	TokenBPoolAddress string          `json:"-" gorm:"type:varchar(64);  index"`
	TokenSwapAddress  string          `json:"-" gorm:"type:varchar(64);  index"`
	LastTransaction   string          `json:"-" gorm:"type:varchar(1024);"`
	Signature         string          `json:"-" gorm:"type:varchar(1024);"`
	PairName          string          `json:"-" gorm:"type:varchar(64);"`
	TokenASymbol      string          `json:"-" gorm:"type:varchar(32);"`
	TokenBSymbol      string          `json:"-" gorm:"type:varchar(32);"`
	TokenADecimal     int             `json:"-" gorm:"type:int2;"`
	TokenBDecimal     int             `json:"-" gorm:"type:int2;"`
	TxNum             uint64          `json:"tx_num" gorm:""`
}

func (*SwapPairCount) TableName() string {
	return "swap_pairs_counts"
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

// JsonString 自定义json gorm byte类型
type JsonString string

func (j *JsonString) MarshalJSON() ([]byte, error) {
	return []byte(*j), nil
}

func (j *JsonString) UnmarshalJSON(data []byte) error {
	*j = JsonString(data)
	return nil
}

func (j *JsonString) Value() (driver.Value, error) {
	return driver.String.ConvertValue(*j)
}

func (j *JsonString) Scan(v interface{}) error {
	*j = JsonString(v.(string))
	return nil
}
