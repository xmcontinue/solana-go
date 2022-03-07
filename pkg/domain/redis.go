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
	key := fmt.Sprintf("swap:vol:count:last24h:%s", accountAddress)
	return RedisKey{key, time.Hour * 1, 0}
}

// SwapCountKey swap account 的锁仓量
func SwapCountKey(swapAddress string) RedisKey {
	key := fmt.Sprintf("swap:tvl:%s", swapAddress)
	return RedisKey{key, 0, 0}
}

// SwapTotalCountKey swap 总统计
func SwapTotalCountKey() RedisKey {
	return RedisKey{"swap:count:total", time.Hour, 0}
}

// AccountSwapVolCountKey 总交易额
// 使用两个accountAddress，当获取swap account时，第二个为空，
// 当获取user account 时，第一个表示useraccount ，第二个表示对应的swap account 地址，因为一个user swap 可能参与多个swap 交易
func AccountSwapVolCountKey(accountAddress1, accountAddress2 string) RedisKey {
	key := fmt.Sprintf("swap:vol:%s:%s", accountAddress1, accountAddress2)
	return RedisKey{key, 0, 0}
}
