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

// AccountSwapVolCountKey 总交易额
func AccountSwapVolCountKey(accountAddress string) RedisKey {
	key := fmt.Sprintf("swap:vol:%s", accountAddress)
	return RedisKey{key, 0, 0}
}
