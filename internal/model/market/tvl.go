package model

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	sq "github.com/Masterminds/squirrel"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QuerySwapTransaction(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTransaction, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.SwapTransaction
		total int64
		err   error
	)
	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil
}

// SumLast24hVol 计算最近24小时内的交易额
func SumLast24hVol(ctx context.Context, typ string, filter ...Filter) (decimal.Decimal, error) {

	var (
		err     error
		db      = rDB(ctx)
		balance []*decimal.Decimal
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).Select("sum(token_a_volume)").Group(typ).Find(&balance).Error; err != nil {
		return decimal.Zero, errors.Wrap(err)
	}

	return decimal.Zero, nil
}

func UpsertSwapTvlCount(ctx context.Context, swapTvlCount *domain.SwapTvlCount) (*domain.SwapTvlCount, error) {
	var (
		after   domain.SwapTvlCount
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": swapTvlCount.LastSwapTransactionID,
			"swap_address":             swapTvlCount.SwapAddress,
			"token_a_address":          swapTvlCount.TokenAAddress,
			"token_b_address":          swapTvlCount.TokenBAddress,
			"token_a_volume":           swapTvlCount.TokenAVolume,
			"token_b_volume":           swapTvlCount.TokenBVolume,
			"token_a_balance":          swapTvlCount.TokenABalance,
			"token_b_balance":          swapTvlCount.TokenBBalance,
			"tvl":                      swapTvlCount.Tvl,
			"vol":                      swapTvlCount.Vol,
			"updated_at":               &now,
			"created_at":               &now,
		}
	)

	sql, args, err := sq.Insert("swap_tvl_counts").SetMap(inserts).Suffix("ON CONFLICT(swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapTvlCount.LastSwapTransactionID).
		Suffix("tvl = ?,", swapTvlCount.Tvl).
		Suffix("vol = ?,", swapTvlCount.Vol).
		Suffix("token_a_volume = ?,", swapTvlCount.TokenAVolume).
		Suffix("token_b_volume = ?,", swapTvlCount.TokenBVolume).
		Suffix("token_a_balance = ?,", swapTvlCount.TokenABalance).
		Suffix("token_b_balance = ?", swapTvlCount.TokenBBalance).
		Suffix("WHERE ").
		Suffix("swap_tvl_counts.last_swap_transaction_id < ?", swapTvlCount.LastSwapTransactionID).
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sql, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func UpsertSwapTvlCountDay(ctx context.Context, swapTvlCount *domain.SwapTvlCountDay, blockDate *time.Time) (*domain.SwapTvlCountDay, error) {
	var (
		after   domain.SwapTvlCountDay
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": swapTvlCount.LastSwapTransactionID,
			"swap_address":             swapTvlCount.SwapAddress,
			"token_a_address":          swapTvlCount.TokenAAddress,
			"token_b_address":          swapTvlCount.TokenBAddress,
			"token_a_volume":           swapTvlCount.TokenAVolume,
			"token_b_volume":           swapTvlCount.TokenBVolume,
			"token_a_balance":          swapTvlCount.TokenABalance,
			"token_b_balance":          swapTvlCount.TokenBBalance,
			"tvl":                      swapTvlCount.Tvl,
			"vol":                      swapTvlCount.Vol,
			"updated_at":               &now,
			"created_at":               &now,
			"date":                     blockDate,
			"tx_num":                   1,
		}
	)

	sql, args, err := sq.Insert("swap_tvl_count_days").SetMap(inserts).Suffix("ON CONFLICT(swap_address,date) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapTvlCount.LastSwapTransactionID).
		Suffix("tvl = ?,", swapTvlCount.Tvl).
		Suffix("vol = swap_tvl_count_days.vol + ?,", swapTvlCount.Vol).
		Suffix("token_a_volume = ?,", swapTvlCount.TokenAVolume).
		Suffix("token_b_volume = ?,", swapTvlCount.TokenBVolume).
		Suffix("token_a_balance = ?,", swapTvlCount.TokenABalance).
		Suffix("token_b_balance = ?,", swapTvlCount.TokenBBalance).
		Suffix("tx_num = swap_tvl_count_days.tx_num + 1").
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sql, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func UpsertUserSwapCount(ctx context.Context, userSwapCount *domain.UserSwapCount) (*domain.UserSwapCount, error) {
	var (
		after   domain.UserSwapCount
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": userSwapCount.LastSwapTransactionID,
			"user_address":             userSwapCount.UserAddress,
			"swap_address":             userSwapCount.SwapAddress,
			"token_a_address":          userSwapCount.TokenAAddress,
			"token_b_address":          userSwapCount.TokenBAddress,
			"user_token_a_volume":      userSwapCount.UserTokenAVolume,
			"user_token_b_volume":      userSwapCount.UserTokenBVolume,
			"user_token_a_balance":     userSwapCount.UserTokenABalance,
			"user_token_b_balance":     userSwapCount.UserTokenBBalance,
			"updated_at":               &now,
			"created_at":               &now,
			"tx_num":                   1,
		}
	)

	sql, args, err := sq.Insert("user_swap_counts").SetMap(inserts).Suffix("ON CONFLICT(user_address,swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", userSwapCount.LastSwapTransactionID).
		Suffix("user_token_a_volume = user_swap_counts.user_token_a_volume + ?,", userSwapCount.UserTokenAVolume).
		Suffix("user_token_b_volume = user_swap_counts.user_token_b_volume + ?,", userSwapCount.UserTokenBVolume).
		Suffix("user_token_a_balance = ?,", userSwapCount.UserTokenABalance).
		Suffix("user_token_b_balance = ?,", userSwapCount.UserTokenBBalance).
		Suffix("tx_num = user_swap_counts.tx_num +1").
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sql, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func UpsertUserSwapCountDay(ctx context.Context, userSwapCount *domain.UserSwapCountDay, blockDate *time.Time) (*domain.UserSwapCountDay, error) {

	var (
		after   domain.UserSwapCountDay
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": userSwapCount.LastSwapTransactionID,
			"user_address":             userSwapCount.UserAddress,
			"swap_address":             userSwapCount.SwapAddress,
			"token_a_address":          userSwapCount.TokenAAddress,
			"token_b_address":          userSwapCount.TokenBAddress,
			"user_token_a_volume":      userSwapCount.UserTokenAVolume,
			"user_token_b_volume":      userSwapCount.UserTokenBVolume,
			"user_token_a_balance":     userSwapCount.UserTokenABalance,
			"user_token_b_balance":     userSwapCount.UserTokenBBalance,
			"updated_at":               &now,
			"created_at":               &now,
			"tx_num":                   1,
			"date":                     blockDate,
		}
	)

	sql, args, err := sq.Insert("user_swap_count_days").SetMap(inserts).Suffix("ON CONFLICT(user_address,swap_address,date) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", userSwapCount.LastSwapTransactionID).
		Suffix("user_token_a_volume = user_swap_count_days.user_token_a_volume + ?,", userSwapCount.UserTokenAVolume).
		Suffix("user_token_b_volume = user_swap_count_days.user_token_b_volume + ?,", userSwapCount.UserTokenBVolume).
		Suffix("user_token_a_balance = ?,", userSwapCount.UserTokenABalance).
		Suffix("user_token_b_balance = ?,", userSwapCount.UserTokenBBalance).
		Suffix("tx_num = user_swap_count_days.tx_num +1").
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sql, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func QueryUserSwapCountDay(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.UserSwapCountDay, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.UserSwapCountDay
		total int64
		err   error
	)
	if err = db.Model(&domain.UserSwapCountDay{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil

}

func GetLastSwapTvlCount(ctx context.Context, filter ...Filter) (*domain.SwapTvlCount, error) {
	var swapTvlCount *domain.SwapTvlCount
	if err := wDB(ctx).Model(&domain.SwapTvlCount{}).Scopes(filter...).Order("id desc").First(&swapTvlCount).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return swapTvlCount, nil
}

func QuerySwapAddressGroupBySwapAddress(ctx context.Context, filter ...Filter) ([]string, error) {
	var (
		address = make([]string, 0)
	)

	if err := wDB(ctx).Model(&domain.SwapTvlCount{}).Select("swap_address").Scopes(filter...).Group("swap_address").Scan(&address).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return address, nil
}

func QueryUserAddressGroupByUserAddress(ctx context.Context, filter ...Filter) ([]string, error) {
	var (
		address = make([]string, 0)
	)

	if err := wDB(ctx).Model(&domain.UserSwapCount{}).Select("user_address").Scopes(filter...).Group("user_address").Scan(&address).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return address, nil
}

func QuerySwapTvlCountDay(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTvlCountDay, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.SwapTvlCountDay
		total int64
		err   error
	)
	if err = db.Model(&domain.SwapTvlCountDay{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil
}

func QueryUserSwapCount(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.UserSwapCount, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.UserSwapCount
		total int64
		err   error
	)
	if err = db.Model(&domain.UserSwapCount{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil
}
