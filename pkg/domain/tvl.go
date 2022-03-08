package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type DateType string

var (
	DateNone       DateType = ""
	DateMin        DateType = "1min"
	DateTwelfth    DateType = "5min" // 5分钟
	DateQuarter    DateType = "15min"
	DateHalfAnHour DateType = "30min"
	DateHour       DateType = "hour"
	DateDay        DateType = "day"
	DateWek        DateType = "wek"
	DateMon        DateType = "mon"
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

type SwapCount struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`                                           // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:swap_count_swap_address_unique_key"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                     // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                     // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                          // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                          // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                         // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                         // swap token b 余额
	TxNum                 int64           `json:"tx_num"`                                                                                       // 交易笔数
}

type SwapCountKLine struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`                                                         // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:swap_count_k_line_date_swap_address_unique_key"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                                   // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                                   // swap token b 地址
	TokenAVolume          decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                        // swap token a 总交易额
	TokenBVolume          decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                        // swap token b 总交易额
	TokenABalance         decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                                       // swap token a 余额
	TokenBBalance         decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                                       // swap token b 余额
	Date                  *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:swap_count_k_line_date_swap_address_unique_key"`          // 统计日期
	TxNum                 int64           `json:"tx_num"`                                                                                                     // 交易笔数
	DateType              DateType        `json:"date_type" gorm:"not null;type:varchar(64);  uniqueIndex:swap_count_k_line_date_swap_address_unique_key"`    // 时间类型（min,quarter,hour,day,wek,mon）
	Open                  decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                  // 统计时间段累的第一个值
	High                  decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                  // 最大值
	Low                   decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                  // 最小值
	Settle                decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                // 结束值
	Avg                   decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                   // 平均值
	TokenAUSD             decimal.Decimal `json:"token_a_usd" gorm:"column:token_a_usd;type:decimal(36,18);default:1"`                                        // swap token a usd价格 TODO
	TokenBUSD             decimal.Decimal `json:"token_b_usd" gorm:"column:token_b_usd;type:decimal(36,18);default:1"`                                        // swap token b usd价格 TODO
	TvlInUsd              decimal.Decimal `json:"tvl_in_usd" gorm:"type:decimal(36,18);"`                                                                     // tvl（总锁仓量，单位usd）TODO
	VolInUsd              decimal.Decimal `json:"vol_in_usd" gorm:"type:decimal(36,18);"`                                                                     // tvl（总交易量，单位usd）TODO
}

type Tvl struct {
	ID            int64        `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt     *time.Time   `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt     *time.Time   `json:"-" gorm:"not null;type:timestamp(6);index"`
	TotalTvlInUsd string       `json:"total_tvl_in_usd" gorm:"type:varchar(32);"`
	TotalVolInUsd string       `json:"total_vol_in_usd" gorm:"type:varchar(32);"`
	TxNum         uint64       `json:"tx_num"`
	CumuTxNum     uint64       `json:"cumu_tx_num"`
	CumuVolInUsd  string       `json:"cumu_vol_in_usd"`
	Pairs         *PairTvlList `json:"pairs" gorm:"type:text;"`
}

type PairTvlList []*PairTvl

type PairTvl struct {
	Name         string `json:"name"`
	TvlInUsd     string `json:"tvl_in_usd"`
	VolInUsd     string `json:"vol_in_usd"`
	TxNum        uint64 `json:"tx_num"`
	Apr          string `json:"apr"`
	SwapAccount  string `json:"swap_account"`
	CumuTxNum    uint64 `json:"cumu_tx_num"`
	CumuVolInUsd string `json:"cumu_vol_in_usd"`
}

type SwapCountKLineVolCount struct {
	TokenAVolume decimal.Decimal `json:"token_a_volume"` // swap token a 总交易额
	TokenBVolume decimal.Decimal `json:"token_b_volume"` // swap token b 总交易额
	TxNum        uint64          `json:"tx_num"`         // swap token 总交易笔数
}

type SwapCountToApi struct {
	TvlInUsd    string                 `json:"tvl_in_usd"`
	VolInUsd24h string                 `json:"vol_in_usd_24h"`
	TxNum24h    uint64                 `json:"tx_num_24h"`
	VolInUsd    string                 `json:"vol_in_usd"`
	TxNum       uint64                 `json:"tx_num"`
	UserNum     int64                  `json:"user_num"`
	TokenNum    int                    `json:"token_num"`
	Pools       []*SwapCountToApiPool  `json:"pools"`
	Tokens      []*SwapCountToApiToken `json:"tokens"`
}

type SwapCountToApiPool struct {
	Name           string           `json:"name"`
	TvlInUsd       string           `json:"tvl_in_usd"`
	VolInUsd24h    string           `json:"vol_in_usd_24h"`
	TxNum24h       uint64           `json:"tx_num_24h"`
	Apr            string           `json:"apr"`
	SwapAccount    string           `json:"swap_account"`
	TxNum          uint64           `json:"tx_num"`
	VolInUsd       string           `json:"vol_in_usd"`
	PriceIntervals []*PriceInterval `json:"price_intervals"`
	Price          string           `json:"price"`
	PriceRate24h   string           `json:"price_rate_24h"`
}

type SwapCountToApiToken struct {
	Name         string `json:"name"`
	TvlInUsd     string `json:"tvl_in_usd"`
	VolInUsd24h  string `json:"vol_in_usd_24h"`
	TxNum24h     uint64 `json:"tx_num_24h"`
	TxNum        uint64 `json:"tx_num"`
	VolInUsd     string `json:"vol_in_usd"`
	Price        string `json:"price"`
	PriceRate24h string `json:"price_rate_24h"`
}

type PriceInterval struct {
	HighPrice string `json:"high_price" mapstructure:"high_price"`
	LowPrice  string `json:"low_price" mapstructure:"low_price"`
}

func (pt *PairTvlList) Value() (driver.Value, error) {
	b, err := json.Marshal(pt)
	if err != nil {
		return nil, err
	}
	return driver.String.ConvertValue(b)
}

func (pt *PairTvlList) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), pt)
}
