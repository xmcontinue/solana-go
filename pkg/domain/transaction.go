package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/util/decimal"
	"github.com/xmcontinue/solana-go/rpc"
)

type SwapTransaction struct {
	ID             int64           `json:"id" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt      *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt      *time.Time      `json:"-" gorm:"not null;index"`
	Signature      string          `json:"signature" gorm:"not null;type:varchar(128);  index; uniqueIndex:swap_transaction_signature_swap_address_unique_key"`         // 交易签名
	Fee            decimal.Decimal `json:"fee" gorm:"type:decimal(36,18)"`                                                                                              // 手续费
	BlockTime      *time.Time      `json:"block_time" gorm:"not null;index"`                                                                                            // 打包时间
	Slot           uint64          `json:"slot"  gorm:"not null"`                                                                                                       // 区块高度
	UserAddress    string          `json:"user_address" gorm:"not null;type:varchar(64);  index"`                                                                       // 用户账户
	InstructionLen uint64          `json:"instruction_len" gorm:"not null;default:0;"`                                                                                  // instruction 第一个data长度
	SwapAddress    string          `json:"swap_address" gorm:"not null;type:varchar(64);  index; uniqueIndex:swap_transaction_signature_swap_address_unique_key;index"` // swap地址
	TokenAAddress  string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                                                    // swap token a 地址
	TokenBAddress  string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                                                    // swap token b 地址
	TokenAVolume   decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                                         // swap token a 总交易额
	TokenBVolume   decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                                         // swap token b 总交易额
	TokenABalance  decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                                                        // swap token a 余额
	TokenBBalance  decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                                                        // swap token b 余额
	TokenAUSD      decimal.Decimal `json:"token_a_usd" gorm:"column:token_a_usd;type:decimal(36,18);default:1"`                                                         // swap token a usd价格
	TokenBUSD      decimal.Decimal `json:"token_b_usd" gorm:"column:token_b_usd;type:decimal(36,18);default:1"`                                                         // swap token b usd价格
	Status         bool            `json:"status"`                                                                                                                      // 交易状态: false-失败，true-成功(废弃)
	TxData         *TxData         `json:"tx_data"               gorm:"type:text;" `                                                                                    // 原数据（json格式）
	TxType         string          `json:"tx_type"  gorm:"type:text"`                                                                                                   // tx 类型
}

func (*SwapTransaction) TableName() string {
	return "swap_transactions"
}

type SwapTransactionV2 struct {
	ID          int64           `json:"id" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt   *time.Time      `json:"-" gorm:"not null"`
	UpdatedAt   *time.Time      `json:"-" gorm:"not null"`
	Signature   string          `json:"signature" gorm:"not null;type:varchar(128); uniqueIndex:swap_transaction_signature_swap_address"` // 交易签名
	SwapAddress string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:swap_transaction_signature_swap_address"`
	UserAddress string          `json:"user_address" gorm:"type:varchar(64)"`
	FeePayer    string          `json:"fee_payer"`
	BlockTime   *time.Time      `json:"block_time" gorm:"not null;type:timestamp(6);index"`                  // 打包时间
	Slot        uint64          `json:"slot"  gorm:"not null"`                                               // 区块高度
	Msg         string          `json:"msg"   gorm:"type:text"`                                              // 日志信息
	TokenAUSD   decimal.Decimal `json:"token_a_usd" gorm:"column:token_a_usd;type:decimal(36,18);default:1"` // swap token a usd价格
	TokenBUSD   decimal.Decimal `json:"token_b_usd" gorm:"column:token_b_usd;type:decimal(36,18);default:1"` // swap token b usd价格
	TxType      string          `json:"tx_type"     gorm:"type:varchar(64)"`
	IsMigrate   bool            `json:"is_migrate"  gorm:"default:false"`
}

func (*SwapTransactionV2) TableName() string {
	return "swap_transaction_v2"
}

type Event struct {
	Name string `json:"name"` // 事件名称
	Data string `json:"data"` // 事件数据
}

type SwapPairBase struct {
	ID                     int64           `json:"id" gorm:"primaryKey;auto_increment;index"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt              *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt              *time.Time      `json:"-" gorm:"not null;index"`
	SwapAddress            string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex"` // swap地址
	TokenAAddress          string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`    // swap token a 地址
	TokenBAddress          string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`    // swap token b 地址
	IsSync                 bool            `json:"is_sync"`                                                     // 是否同步至起始区块
	StartSignature         string          `json:"start_signature" gorm:"not null;type:varchar(128)"`           // 当前起始签名
	EndSignature           string          `json:"end_signature" gorm:"not null;type:varchar(128)"`             // 当前最新签名
	FailedTxNum            uint64          `json:"failed_tx_num" gorm:"default:0"`                              // 失败交易笔数
	TotalTxNum             uint64          `json:"total_tx_num" gorm:"default:0"`                               // 总交易笔数 TODO 待开发，由tx解析后统计
	TotalVol               decimal.Decimal `json:"total_vol" gorm:"type:decimal(36,18);default:0"`              // 总交易量 TODO 待开发，由tx解析后统计
	TokenNum               uint64          `json:"token_num" gorm:"default:0"`                                  // token数量 TODO 待开发，由配置文件中解析统计
	UserNum                uint64          `json:"user_num" gorm:"default:0"`                                   // 用户数量 TODO 待开发，由用户总统计表中统计
	SyncUtilFinished       bool            `json:"sync_util_finished" gorm:"default:false"`                     // 是否重新同步user_address和类型完成
	MigrateID              int64           `json:"migrate_id" gorm:"default:0"`                                 // 迁移进度
	PairPriceMigrateID     int64           `json:"pair_price_migrate_id" gorm:"default:0"`                      // 价格表迁移进度
	IsPriceMigrateFinished bool            `json:"is_price_migrate_finished" gorm:"default:false"`              // 价格是否同步到最新
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
