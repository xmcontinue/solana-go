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

// SwapVolCountLast24HKey 最近24小时的总的交易额
func SwapVolCountLast24HKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("crema:swap:vol:count:last24h:%s", accountAddress)
	return RedisKey{key, time.Hour * 1, 0}
}

// SwapTvlCountKey swap account 的锁仓量
func SwapTvlCountKey(swapAddress string) RedisKey {
	key := fmt.Sprintf("crema:swap:tvl:count:%s", swapAddress)
	return RedisKey{key, 0, 0}
}

// SwapTotalCountKey swap 总统计的redis key
func SwapTotalCountKey() RedisKey {
	return RedisKey{"crema:swap:count:total", time.Hour, 0}
}

// AccountSwapVolCountKey 总交易额
// 使用两个accountAddress，当获取swap account时，第二个为空，
// 当获取user account 时，第一个表示useraccount ，第二个表示对应的swap account 地址，因为一个user swap 可能参与多个swap 交易
func AccountSwapVolCountKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("crema:swap:vol:%s", accountAddress)
	return RedisKey{key, 0, 0}
}

func KLineKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("crema:kline:swap:count:%s:%s", dateType, swapAccount)
}

func HistogramKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("crema:histogram:swap:count:%s:%s", dateType, swapAccount)
}

func TotalHistogramKey(dateType DateType) string {
	return fmt.Sprintf("crema:histogram:total:swap:count:%s", dateType)
}
