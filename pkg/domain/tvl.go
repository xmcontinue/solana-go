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
	CreatedAt        *time.Time `json:"-" gorm:"not null;index"`
	UpdatedAt        *time.Time `json:"-" gorm:"not null;index"`
	RpcURL           string     `json:"rpc_url"  gorm:"not null;type:varchar(128)"` // Rpc地址
	LastSlot         int64      `json:"last_slot" gorm:"not null"`                  // 最新区块高度
	FailedRequestNum int64      `json:"failed_request_num" gorm:"not null"`         // 请求失败次数
	Enable           bool       `json:"enable" gorm:"default:false"`                // 是否已启用，默认不启用
}

type SwapCount struct {
	ID                     int64           `json:"-" gorm:"primaryKey;AUTO_INCREMENT"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt              *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt              *time.Time      `json:"-" gorm:"not null;index"`
	LastSwapTransactionID  int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`        // 最后同步的transaction id
	SwapAddress            string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex"` // swap地址
	TokenAAddress          string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`  // swap token a 地址
	TokenBAddress          string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`  // swap token b 地址
	TokenAVolume           decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`       // swap token a 总交易额
	TokenBVolume           decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`       // swap token b 总交易额
	TokenABalance          decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`      // swap token a 余额
	TokenBBalance          decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`      // swap token b 余额
	TxNum                  int64           `json:"tx_num"`                                                    // 交易笔数
	MigrateSwapContKLineID int64           `json:"migrate_swap_cont_k_line_id" gorm:"default:0"`              // swapCountKline 迁移进度
}

type SwapCountMigrate struct {
	ID                     int64      `json:"-" gorm:"primaryKey;AUTO_INCREMENT"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt              *time.Time `json:"-" gorm:"not null;index"`
	UpdatedAt              *time.Time `json:"-" gorm:"not null;index"`
	SwapAddress            string     `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex"` // swap地址
	MigrateSwapContKLineID int64      `json:"migrate_swap_cont_k_line_id" gorm:"default:0"`              // swapCountKline 迁移进度
}

type SwapCountKLine struct {
	ID                       int64           `json:"-" gorm:"primaryKey;auto_increment; Index:SwapCountKLine_id_token_ausd_for_contract_index"`                                                                                                                                            // 自增主键，自增主键不能有任何业务含义。
	CreatedAt                *time.Time      `json:"-" gorm:"not null"`                                                                                                                                                                                                                    // 始终不会用到这个索引
	UpdatedAt                *time.Time      `json:"-" gorm:"not null"`                                                                                                                                                                                                                    // 始终不会用到这个索引
	LastSwapTransactionID    int64           `json:"last_swap_transaction_id" gorm:"not null;default:0; index"`                                                                                                                                                                            // 最后同步的transaction id
	SwapAddress              string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:swap_count_k_line_date_swap_address_unique_key; index;index:swap_count_k_line_date_type_swap_address_date,priority:2;index:swap_count_k_line_swap_address_date_type_date"` // swap地址
	TokenAAddress            string          `json:"token_a_address" gorm:"not null;type:varchar(64);"`                                                                                                                                                                                    // swap token a 地址
	TokenBAddress            string          `json:"token_b_address" gorm:"not null;type:varchar(64);"`                                                                                                                                                                                    // swap token b 地址
	TokenAVolume             decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                  // swap token a 总交易额（发起量）
	TokenBVolume             decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                  // swap token b 总交易额（发起量）
	TokenAQuoteVolume        decimal.Decimal `json:"token_a_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                            // swap token a 获得量
	TokenBQuoteVolume        decimal.Decimal `json:"token_b_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                            // swap token b 获得量
	TokenABalance            decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                 // swap token a 余额
	TokenBBalance            decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                 // swap token b 余额
	TokenARefAmount          decimal.Decimal `json:"token_a_ref_amount" gorm:"type:decimal(36,18);default:0"`
	TokenAFeeAmount          decimal.Decimal `json:"token_a_fee_amount" gorm:"type:decimal(36,18);default:0"`
	TokenAProtocolAmount     decimal.Decimal `json:"token_a_protocol_amount" gorm:"type:decimal(36,18);default:0"` // swap token b 余额
	TokenBRefAmount          decimal.Decimal `json:"token_b_ref_amount" gorm:"type:decimal(36,18);default:0"`
	TokenBFeeAmount          decimal.Decimal `json:"token_b_fee_amount" gorm:"type:decimal(36,18);default:0"`
	TokenBProtocolAmount     decimal.Decimal `json:"token_b_protocol_amount" gorm:"type:decimal(36,18);default:0"`
	TokenASymbol             string          `json:"token_a_symbol"      gorm:"not null;type:varchar(64);  index:swap_count_k_line_token_a_symbol_date_type_date"` // token A 币种符号
	TokenBSymbol             string          `json:"token_b_symbol"      gorm:"not null;type:varchar(64);  index:swap_count_k_line_token_b_symbol_date_type_date"`
	DateType                 DateType        `json:"date_type"  gorm:"not null;type:varchar(64);  uniqueIndex:swap_count_k_line_date_swap_address_unique_key; index:swap_count_k_line_token_a_symbol_date_type_date;index:swap_count_k_line_token_b_symbol_date_type_date;index:swap_count_k_line_date_type_swap_address_date,priority:1;index:swap_count_k_line_swap_address_date_type_date"` // 时间类型（min,quarter,hour,day,wek,mon）
	Date                     *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:swap_count_k_line_date_swap_address_unique_key; index:swap_count_k_line_token_a_symbol_date_type_date;index:swap_count_k_line_token_b_symbol_date_type_date;index;index:swap_count_k_line_date_type_swap_address_date,priority:3;index:swap_count_k_line_swap_address_date_type_date"`  // 统计日期
	TxNum                    int64           `json:"tx_num"`                                                                                                                                                                                                                                                                                                                                   // 交易笔数
	Open                     decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                                                                                                                                // 统计时间段累的第一个值
	High                     decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                                                                                                                                // 最大值
	Low                      decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                                                                                                                                // 最小值
	Settle                   decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                                                                                                                              // 结束值
	Avg                      decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                                                                                                                                                                                                                                                 // 平均值 	// token B 币种符号
	TokenAUSD                decimal.Decimal `json:"token_a_usd" gorm:"column:token_a_usd;type:decimal(36,18);default:1"`                                                                                                                                                                                                                                                                      // swap token a usd价格
	TokenBUSD                decimal.Decimal `json:"token_b_usd" gorm:"column:token_b_usd;type:decimal(36,18);default:1"`                                                                                                                                                                                                                                                                      // swap token b usd价格
	TvlInUsd                 decimal.Decimal `json:"tvl_in_usd" gorm:"type:decimal(36,18);"`                                                                                                                                                                                                                                                                                                   // tvl（总锁仓量，单位usd）
	VolInUsd                 decimal.Decimal `json:"vol_in_usd" gorm:"type:decimal(36,18);"`                                                                                                                                                                                                                                                                                                   // tvl（总交易量，单位usd）
	VolInUsdForContract      decimal.Decimal `json:"vol_in_usd_for_contract" gorm:"type:decimal(36,18);"`                                                                                                                                                                                                                                                                                      // 实时计算总交易量（总交易量，单位usd）
	TokenAUSDForContract     decimal.Decimal `json:"token_ausd_for_contract" gorm:"column:token_ausd_for_contract;type:decimal(36,18);default:0; Index:SwapCountKLine_id_token_ausd_for_contract_index"`                                                                                                                                                                                       // swap token a usd价格(合约内部价格)
	TokenBUSDForContract     decimal.Decimal `json:"token_busd_for_contract" gorm:"column:token_busd_for_contract;type:decimal(36,18);default:0"`                                                                                                                                                                                                                                              // swap token b usd价格(合约内部价格)
	MaxBlockTimeWithDateType *time.Time      `json:"-"  gorm:"type:timestamp(6)"`
	IsMigrate                bool            `json:"is_migrate" gorm:"default:false"`
}

func (*SwapCountKLine) TableName() string {
	return "swap_count_k_lines"
}

type Tvl struct {
	ID            int64        `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt     *time.Time   `json:"-" gorm:"not null;index"`
	UpdatedAt     *time.Time   `json:"-" gorm:"not null;index"`
	TotalTvlInUsd string       `json:"total_tvl_in_usd" gorm:"type:text;"`
	TotalVolInUsd string       `json:"total_vol_in_usd" gorm:"type:text;"`
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
	TokenAVolume            decimal.Decimal `json:"token_a_volume"`               // swap token a 总交易额(发起方)
	TokenBVolume            decimal.Decimal `json:"token_b_volume"`               // swap token b 总交易额(发起方)
	TokenAQuoteVolume       decimal.Decimal `json:"token_a_quote_volume"`         // swap token a 交易额(获得方)
	TokenBQuoteVolume       decimal.Decimal `json:"token_b_quote_volume"`         // swap token b 交易额(获得方)
	TokenAVolumeForUsd      decimal.Decimal `json:"token_a_volume_for_usd"`       // swap token a 总交易额(发起方)(USD)
	TokenBVolumeForUsd      decimal.Decimal `json:"token_b_volume_for_usd"`       // swap token b 总交易额(发起方)(USD)
	TokenAQuoteVolumeForUsd decimal.Decimal `json:"token_a_quote_volume_for_usd"` // swap token a 交易额(获得方)(USD)
	TokenBQuoteVolumeForUsd decimal.Decimal `json:"token_b_quote_volume_for_usd"` // swap token b 交易额(获得方)(USD)
	VolInUsdForContract     decimal.Decimal `json:"vol_in_usd_for_contract"`      // 累计交易额
	FeeAmount               decimal.Decimal `json:"fee_amount"`                   // swap token b 交易额(获得方)(USD)
	RefAmount               decimal.Decimal `json:"ref_amount"`                   // 渠道商获取的总fee
	ProtocolAmount          decimal.Decimal `json:"protocol_amount"`              // 池子获取的总fee
	TxNum                   uint64          `json:"tx_num"`                       // swap token 总交易笔数
	DayNum                  uint64          `json:"day_num"`
}

type SwapCountListInfo struct {
	TokenAUSDForContract decimal.Decimal `json:"token_ausd_for_contract" gorm:"column:token_ausd_for_contract;type:decimal(36,18);default:0; Index: SwapCountKLine_id_token_ausd_for_contract_index"` // swap token a usd价格(合约内部价格)
	TokenBUSDForContract decimal.Decimal `json:"token_busd_for_contract" gorm:"column:token_busd_for_contract;type:decimal(36,18);default:0"`                                                         // swap token b usd价格(合约内部价格)
	TokenAVolume         decimal.Decimal `json:"token_a_volume"`                                                                                                                                      // swap token a 总交易额(发起方)
	TokenBVolume         decimal.Decimal `json:"token_b_volume"`                                                                                                                                      // swap token b 总交易额(发起方)
	TxNum                uint64          `json:"tx_num"`                                                                                                                                              // swap token 总交易笔数
	Date                 *time.Time      `json:"date"`
	SwapAddress          string          `json:"swap_address"`
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
	Name string `json:"name"`
	// Test24hFeeAmount            string         `json:"test_24_h_fee_amount"` // 仅测试使用
	// Test7dFeeAmount             string         `json:"test_7_d_fee_amount"`
	// Test7dNum                   uint64         `json:"test_7_d_num"`
	// Test30dNum                  uint64         `json:"test_30_d_num"`
	// Test30dFeeAmount            string         `json:"test_30_d_fee_amount"`
	// Test24hRefAmount            string         `json:"test_24h_ref_amount"`
	// Test24hProAmount            string         `json:"test_24h_pro_amount"`
	// Test7dRefAmount             string         `json:"test_7d_ref_amount"`
	// Test7dProAmount             string         `json:"test_7d_pro_amount"`
	// Test30dRefAmount            string         `json:"test_30d_ref_amount"`
	// Test30dProAmount            string         `json:"test_30d_pro_amount"`
	TvlInUsd                    string         `json:"tvl_in_usd"`
	VolInUsd24h                 string         `json:"vol_in_usd_24h"`
	TxNum24h                    uint64         `json:"tx_num_24h"`
	Apr                         string         `json:"apr"`
	Fee                         string         `json:"fee"`
	SwapAccount                 string         `json:"swap_account"`
	TokenAReserves              string         `json:"token_a_reserves"`
	TokenBReserves              string         `json:"token_b_reserves"`
	TxNum                       uint64         `json:"tx_num"`
	VolInUsd                    string         `json:"vol_in_usd"`
	PriceInterval               *PriceInterval `json:"price_interval"`
	Price                       string         `json:"price"`
	PriceRate24h                string         `json:"price_rate_24h"`
	VolumeInTokenA24h           string         `json:"volume_in_tokenA_24h"`
	VolumeInTokenB24h           string         `json:"volume_in_tokenB_24h"`
	VolumeInTokenA24hUnilateral string         `json:"volume_in_tokenA_24h_unilateral"`
	VolumeInTokenB24hUnilateral string         `json:"volume_in_tokenB_24h_unilateral"`
	TokenAAddress               string         `json:"token_a_address"`
	TokenBAddress               string         `json:"token_b_address"`
	Version                     string         `json:"version"`
	Apr24h                      string         `json:"apr_24h"`
	Apr7Day                     string         `json:"apr_7day"`
	Apr30Day                    string         `json:"apr_30day"`
	RewarderApr                 string         `json:"rewarder_apr"`
	Fee7D                       string         `json:"fee_7_d"`
	Fee30D                      string         `json:"fee_30_d"`
	Apr7DayCount                int64          `json:"apr_7_day_count"`
	Apr30DayCount               int64          `json:"apr_30_day_count"`
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
	UpperPrice string `json:"upper_price" mapstructure:"upper_price"`
	LowerPrice string `json:"lower_price" mapstructure:"lower_price"`
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

type ShardingKeyAndTableName struct {
	ID               int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt        *time.Time `json:"-" gorm:"not null;index"`
	UpdatedAt        *time.Time `json:"-" gorm:"not null;index"`
	ShardingKeyValue string     `json:"sharding_key_value" gorm:"unique"`
	Suffix           int        `json:"suffix"             gorm:"unique"`
}
