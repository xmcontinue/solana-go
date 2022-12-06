package domain

import (
	"fmt"
	"time"
)

type RedisKey struct {
	Key     string
	Timeout time.Duration
	Size    int64
}

var publicPrefix = "crema"

// SwapVolCountLast24HKey 最近24小时的总的交易额
func SwapVolCountLast24HKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("%s:swap:vol:count:last24h:%s", publicPrefix, accountAddress)
	return RedisKey{key, 0, 0}
}

// SwapTvlCountKey swap account 的锁仓量
func SwapTvlCountKey(swapAddress string) RedisKey {
	key := fmt.Sprintf("%s:swap:tvl:%s", publicPrefix, swapAddress)
	return RedisKey{key, 0, 0}
}

// SwapTotalCountKey swap 总统计的redis key
func SwapTotalCountKey() RedisKey {
	return RedisKey{fmt.Sprintf("%s:swap:count:total", publicPrefix), 0, 0}
}

func SwapTotalCountKeyWithSharding() RedisKey {
	return RedisKey{fmt.Sprintf("%s:swap:count:total:sharding", publicPrefix), 0, 0}
}

// LastSwapTransactionID 如果有新增的表，则新增redis key ，用以判断当前表同步数据位置，且LastSwapTransactionID为截止id
func LastSwapTransactionID() RedisKey {
	return RedisKey{fmt.Sprintf("%s:swap:transaction:last:id", publicPrefix), time.Hour, 0}
}

func SwapStatusKey() RedisKey {
	return RedisKey{fmt.Sprintf("%s:swap:pause:status", publicPrefix), 0, 0}
}

// AccountSwapVolCountKey 总交易额
// 使用两个accountAddress，当获取swap account时，第二个为空，
// 当获取user account 时，第一个表示useraccount ，第二个表示对应的swap account 地址，因为一个user swap 可能参与多个swap 交易
func AccountSwapVolCountKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("%s:swap:vol:%s", publicPrefix, accountAddress)
	return RedisKey{key, 0, 0}
}

func KLineKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("%s:swap:kline:count:%s:%s", publicPrefix, dateType, swapAccount)
}

func TokenKey(swapAccount string) string {
	return fmt.Sprintf("%s:swap:token:count:%s", publicPrefix, swapAccount)
}

func HistogramKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("%s:histogram:swap:count:%s:%s", publicPrefix, dateType, swapAccount)
}

func HistogramKeySharding(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("%s:histogram:swap:count:sharding:%s:%s", publicPrefix, dateType, swapAccount)
}

func TotalHistogramKey(dateType DateType) string {
	return fmt.Sprintf("%s:histogram:total:swap:count:%s", publicPrefix, dateType)
}

func TotalHistogramKeySharding(dateType DateType) string {
	return fmt.Sprintf("%s:histogram:total:swap:count:sharding:%s", publicPrefix, dateType)
}

func SetPublicPrefix(key string) {
	publicPrefix = key
}

func GetPublicPrefix() string {
	return publicPrefix
}
