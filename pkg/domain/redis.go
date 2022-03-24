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

const publicPrefix = "crema:"

// SwapVolCountLast24HKey 最近24小时的总的交易额
func SwapVolCountLast24HKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("%sswap:vol:count:last24h:%s", publicPrefix, accountAddress)
	return RedisKey{key, time.Hour * 1, 0}
}

// SwapTvlCountKey swap account 的锁仓量
func SwapTvlCountKey(swapAddress string) RedisKey {
	key := fmt.Sprintf("%sswap:tvl:%s", publicPrefix, swapAddress)
	return RedisKey{key, 0, 0}
}

// SwapTotalCountKey swap 总统计的redis key
func SwapTotalCountKey() RedisKey {
	return RedisKey{fmt.Sprintf("%sswap:count:total", publicPrefix), time.Hour, 0}
}

// AccountSwapVolCountKey 总交易额
// 使用两个accountAddress，当获取swap account时，第二个为空，
// 当获取user account 时，第一个表示useraccount ，第二个表示对应的swap account 地址，因为一个user swap 可能参与多个swap 交易
func AccountSwapVolCountKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("%sswap:vol:%s", publicPrefix, accountAddress)
	return RedisKey{key, 0, 0}
}

func KLineKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("%sswap:kline:count:%s:%s", publicPrefix, dateType, swapAccount)
}

func TokenKey(swapAccount string) string {
	return fmt.Sprintf("%sswap:token:count:%s", publicPrefix, swapAccount)
}

func HistogramKey(dateType DateType, swapAccount string) string {
	return fmt.Sprintf("%shistogram:swap:count:%s:%s", publicPrefix, dateType, swapAccount)
}

func TotalHistogramKey(dateType DateType) string {
	return fmt.Sprintf("%shistogram:total:swap:count:%s", publicPrefix, dateType)
}
