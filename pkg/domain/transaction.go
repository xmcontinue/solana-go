package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"
)

type SwapTransaction struct {
	ID             int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt      *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt      *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	Signature      string          `json:"signature" gorm:"not null;type:varchar(128);  index"`      // 交易签名
	Fee            decimal.Decimal `json:"fee" gorm:"type:decimal(36,18)"`                           // 手续费
	BlockTime      *time.Time      `json:"block_time" gorm:"not null;type:timestamp(6)"`             // 打包时间
	Slot           uint64          `json:"slot"  gorm:"not null"`                                    // 区块高度
	UserAddress    string          `json:"user_address" gorm:"not null;type:varchar(64);  index"`    // 用户账户
	InstructionLen uint64          `json:"instruction_len" gorm:"not null;default:0;"`               // instruction 第一个data长度
	SwapAddress    string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress  string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress  string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenAVolume   decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`      // swap token a 总交易额
	TokenBVolume   decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`      // swap token b 总交易额
	TokenABalance  decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`     // swap token a 余额
	TokenBBalance  decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`     // swap token b 余额
	Status         bool            `json:"status"`                                                   // 交易状态: 0-失败，1-成功
	TxData         *TxData         `json:"-"               gorm:"type:text;" `                       // 原数据（json格式）
}

type SwapPairBase struct {
	gorm.Model
	SwapAddress    string          `json:"swap_address" gorm:"not null;type:varchar(64);  index"`    // swap地址
	TokenAAddress  string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress  string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	IsSync         bool            `json:"is_sync"`                                                  // 是否同步至起始区块
	StartSignature string          `json:"start_signature" gorm:"not null;type:varchar(128)"`        // 当前起始签名
	EndSignature   string          `json:"end_signature" gorm:"not null;type:varchar(128)"`          // 当前最新签名
	TotalTxNum     uint64          `json:"total_tx_num" gorm:"default:0"`                            // 总交易笔数
	TotalVol       decimal.Decimal `json:"total_vol" gorm:"type:decimal(36,18);default:0"`           // 总交易量
}

type SumVol struct {
	TxNum          uint64          `gorm:"tx_num"`            // 总交易笔数
	TokenATotalVol decimal.Decimal `gorm:"token_a_total_vol"` // tokenA总交易量
	TokenBTotalVol decimal.Decimal `gorm:"token_b_total_vol"` // tokenB总交易量
}

// TxData 自定义tx原始数据类型
type TxData rpc.GetTransactionResult

func (tx *TxData) Value() (driver.Value, error) {
	b, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return driver.String.ConvertValue(b)
}

func (tx *TxData) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), tx)
}
