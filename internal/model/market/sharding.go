package model

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"gorm.io/sharding"

	"git.cplus.link/crema/backend/pkg/domain"
)

var shardingConfig sharding.Config //分表配置

// ShardingKeyAndTableName 分表的表名和分表建之间的关系
type ShardingKeyAndTableName struct {
	ID               int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt        *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt        *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	ShardingKeyValue string     `json:"sharding_key_value" gorm:"unique"`
	Suffix           int        `json:"suffix"             gorm:"unique"`
}

func QueryShardings(ctx context.Context, limit, offset int, filter ...Filter) ([]*ShardingKeyAndTableName, error) {
	var (
		db    = rDB(ctx)
		err   error
		lists []*ShardingKeyAndTableName
	)

	if err = db.Model(&ShardingKeyAndTableName{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&lists).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return lists, nil
}

func CreateShardings(ctx context.Context, shardingKeyAndTableName *ShardingKeyAndTableName) error {
	if err := wDB(ctx).Create(shardingKeyAndTableName).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// ShardingKeyValueToTableIndex 所有分表共用一个map
var ShardingKeyValueToTableIndex map[string]int

func initShardingKeyValue(shardingValues []string) error {
	if err := dbWPool.NewConn().AutoMigrate(
		&ShardingKeyAndTableName{},
	); err != nil {
		return errors.Wrap(err)
	}

	// 先初始化一张表用于存储表名和分表键之间的对应关系
	ShardingKeyValueToTableIndex = make(map[string]int, len(shardingValues))
	maxIndex := -1
	shardings, err := QueryShardings(context.Background(), 0, 0, OrderFilter("suffix asc"))
	if err != nil {
		return errors.Wrap(err)
	}

	if len(shardings) != 0 {
		maxIndex = shardings[len(shardings)-1].Suffix
	}

	// 先获取数据库里面的映射关系
	for _, v := range shardings {
		ShardingKeyValueToTableIndex[v.ShardingKeyValue] = v.Suffix
	}

	// 创建新的分表键值对
	for _, value := range shardingValues {
		if _, ok := ShardingKeyValueToTableIndex[value]; !ok {
			maxIndex++
			ShardingKeyValueToTableIndex[value] = maxIndex
			// 保存已有映射关系
			err = CreateShardings(context.Background(), &ShardingKeyAndTableName{
				ShardingKeyValue: value,
				Suffix:           maxIndex,
			})
			if err != nil {
				return errors.Wrap(err)
			}

		}
	}

	return nil
}

func getTableFullName(table string, swapAddress string) (string, error) {
	suffix, err := shardingConfig.ShardingAlgorithm(swapAddress)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return table + suffix, nil
}

func autoMigrateWithSharding(shardingValues []string) error {

	// 自定义分表算法
	shardingAlgorithm := func(columnValue interface{}) (suffix string, err error) {
		if uid, ok := columnValue.(string); ok {
			k, ok := ShardingKeyValueToTableIndex[uid]
			if !ok {
				return "", errors.New("invalid swap_address")
			}
			return fmt.Sprintf("_%02d", k), nil
		}
		return "", errors.New("invalid swap_address")
	}

	shardingSuffixs := func() (suffixs []string) {
		for _, v := range ShardingKeyValueToTableIndex {
			suffixs = append(suffixs, fmt.Sprintf("_%02d", v))
		}
		return suffixs
	}

	shardingConfig = sharding.Config{
		DoubleWrite:         false, // todo 删除
		ShardingKey:         "swap_address",
		NumberOfShards:      uint(len(shardingValues)),
		PrimaryKeyGenerator: sharding.PKSnowflake,
		ShardingAlgorithm:   shardingAlgorithm,
		ShardingSuffixs:     shardingSuffixs,
	}

	shardingTables := []interface{}{
		&domain.SwapCountKLine{},
		&domain.SwapPairPriceKLine{},
		&domain.SwapTransactionV2{},
	}
	// 添加分表
	dbRPool.WithSharding(shardingConfig,
		shardingTables...,
	)

	// 迁移
	err := wDB(context.Background()).AutoMigrate(shardingTables...)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func createIndex() error {
	err := createSwapContKlineIndex()
	if err != nil {
		return errors.Wrap(err)
	}

	err = createSwapPairPriceKlineIndex()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func createSwapContKlineIndex() error {
	masterTableName := "swap_count_k_lines"
	// 创建索引
	for _, v := range ShardingKeyValueToTableIndex {
		table := fmt.Sprintf("%s_%02d", masterTableName, v)
		tx := wDB(context.Background()).Exec(
			" CREATE UNIQUE INDEX IF NOT EXISTS " + table + "_date_swap_address_unique_key ON " + table + " (\"swap_address\",\"date_type\",\"date\");",
		)
		if tx.Error != nil {
			return errors.Wrap(tx.Error)
		}

	}
	return nil
}

func createSwapPairPriceKlineIndex() error {
	masterTableName := "swap_pair_price_k_lines"
	// 创建索引
	for _, v := range ShardingKeyValueToTableIndex {
		table := fmt.Sprintf("%s_%02d", masterTableName, v)
		tx := wDB(context.Background()).Exec(
			" CREATE UNIQUE INDEX IF NOT EXISTS " + table + "_date_swap_address_unique_key ON " + table + " (\"swap_address\",\"date_type\",\"date\");",
		)
		if tx.Error != nil {
			return errors.Wrap(tx.Error)
		}

	}
	return nil
}

func InitWithSharding(shardingValue []string) error {
	var rErr error

	// 创建swap_account 和表名后缀的映射关系
	if err := initShardingKeyValue(shardingValue); err != nil {
		rErr = errors.Wrapf(err, "auto migrate")
	}

	// 分表自动迁移
	if err := autoMigrateWithSharding(shardingValue); err != nil {
		rErr = errors.Wrapf(err, "auto migrate")
	}

	// 分表不会创建部分索引，必须手动创建
	if err := createIndex(); err != nil {
		rErr = errors.Wrapf(err, "auto migrate")
	}

	return rErr
}
