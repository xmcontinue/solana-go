package model

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	sq "github.com/Masterminds/squirrel"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QuerySwapTransactions(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTransaction, error) {
	var (
		db   = rDB(ctx)
		list []*domain.SwapTransaction

		err error
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return list, nil
}

type SwapVol struct {
	SwapAddress   string          `json:"swap_address"`
	TokenAVolume  decimal.Decimal `json:"token_a_volume"`
	TokenBVolume  decimal.Decimal `json:"token_b_volume"`
	TokenAAddress string          `json:"token_a_address"`
	TokenBAddress string          `json:"token_b_address"`
	Vol           decimal.Decimal `json:"vol"`
	Num           int             `json:"num"`
}

// SumSwapAccountLast24Vol 最近24小时 swap account 的总交易额
func SumSwapAccountLast24Vol(ctx context.Context, filter ...Filter) ([]*SwapVol, error) {
	var (
		err     error
		db      = rDB(ctx)
		swapVol []*SwapVol
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).
		Select("sum(token_a_volume) as token_a_volume,sum(token_b_volume) as token_b_volume,count(*) as num,swap_address,token_a_address,token_b_address").
		Group("swap_address,token_a_address,token_b_address").Find(&swapVol).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapVol, nil
}

type UserSwapVol struct {
	UserAddress   string          `json:"user_address"`
	TokenAVolume  decimal.Decimal `json:"token_a_volume"`
	TokenBVolume  decimal.Decimal `json:"token_b_volume"`
	SwapAddress   string          `json:"swap_address"`
	TokenAAddress string          `json:"token_a_address"`
	TokenBAddress string          `json:"token_b_address"`
	Vol           decimal.Decimal `json:"vol"`
	Num           int             `json:"num"`
}

// SumUserSwapAccountLast24Vol 最近24小时 user account 的总交易额
func SumUserSwapAccountLast24Vol(ctx context.Context, filter ...Filter) ([]*UserSwapVol, error) {
	var (
		err     error
		db      = rDB(ctx)
		swapVol []*UserSwapVol
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).
		Select("sum(token_a_volume) as token_a_volume,sum(token_b_volume) as token_b_volume,count(*) as num,user_address,swap_address,token_a_address,token_b_address").
		Group("user_address,swap_address,token_a_address,token_b_address").Find(&swapVol).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapVol, nil
}

func UpsertSwapCount(ctx context.Context, swapCount *domain.SwapCount) (*domain.SwapCount, error) {
	var (
		after   domain.SwapCount
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": swapCount.LastSwapTransactionID,
			"swap_address":             swapCount.SwapAddress,
			"token_a_address":          swapCount.TokenAAddress,
			"token_b_address":          swapCount.TokenBAddress,
			"token_a_volume":           swapCount.TokenAVolume.Abs(),
			"token_b_volume":           swapCount.TokenBVolume.Abs(),
			"token_a_balance":          swapCount.TokenABalance,
			"token_b_balance":          swapCount.TokenBBalance,
			"updated_at":               &now,
			"created_at":               &now,
			"tx_num":                   1,
		}
	)

	sql, args, err := sq.Insert("swap_counts").SetMap(inserts).Suffix("ON CONFLICT(swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapCount.LastSwapTransactionID).
		Suffix("token_a_volume = swap_counts.token_a_volume + ?,", swapCount.TokenAVolume.Abs()).
		Suffix("token_b_volume = swap_counts.token_b_volume + ?,", swapCount.TokenBVolume.Abs()).
		Suffix("token_a_balance = ?,", swapCount.TokenABalance).
		Suffix("token_b_balance = ?,", swapCount.TokenBBalance).
		Suffix("tx_num = swap_counts.tx_num + 1").
		Suffix("WHERE ").
		Suffix("swap_counts.last_swap_transaction_id <= ?", swapCount.LastSwapTransactionID).
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

func UpsertSwapCountKLine(ctx context.Context, swapCount *domain.SwapCountKLine, blockDate *time.Time) (*domain.SwapCountKLine, error) {
	var (
		after   domain.SwapCountKLine
		now     = time.Now().UTC()
		avgFmt  string
		inserts = map[string]interface{}{
			"last_swap_transaction_id": swapCount.LastSwapTransactionID,
			"swap_address":             swapCount.SwapAddress,
			"token_a_address":          swapCount.TokenAAddress,
			"token_b_address":          swapCount.TokenBAddress,
			"token_a_volume":           swapCount.TokenAVolume.Abs(),
			"token_b_volume":           swapCount.TokenBVolume.Abs(),
			"token_a_balance":          swapCount.TokenABalance,
			"token_b_balance":          swapCount.TokenBBalance,
			"date_type":                swapCount.DateType,
			"open":                     swapCount.Open,
			"high":                     swapCount.High,
			"low":                      swapCount.Low,
			"settle":                   swapCount.Settle,
			"avg":                      swapCount.Avg,
			"updated_at":               &now,
			"created_at":               &now,
			"date":                     blockDate,
			"tx_num":                   1,
			"token_a_usd":              swapCount.TokenAUSD,
			"token_b_usd":              swapCount.TokenBUSD,
			"tvl_in_usd":               swapCount.TvlInUsd,
			"vol_in_usd":               swapCount.VolInUsd,
		}
	)

	// 除了domain.DateMin 类型，其他的都是根据前一个类型求平均值
	if swapCount.DateType == domain.DateMin {
		avgFmt = "avg = (swap_count_k_lines.avg * swap_count_k_lines.tx_num + ? )/(swap_count_k_lines.tx_num+ 1)"
	} else {
		avgFmt = "avg = ?"
	}

	sql, args, err := sq.Insert("swap_count_k_lines").SetMap(inserts).Suffix("ON CONFLICT(swap_address,date,date_type) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapCount.LastSwapTransactionID).
		Suffix("token_a_volume = swap_count_k_lines.token_a_volume + ?,", swapCount.TokenAVolume.Abs()).
		Suffix("token_b_volume = swap_count_k_lines.token_b_volume + ?,", swapCount.TokenBVolume.Abs()).
		Suffix("token_a_balance = ?,", swapCount.TokenABalance).
		Suffix("token_b_balance = ?,", swapCount.TokenBBalance).
		Suffix("high = ?,", swapCount.High).
		Suffix("low = ?,", swapCount.Low).
		Suffix("settle = ?,", swapCount.Settle).
		Suffix("token_a_usd = ?,", swapCount.TokenAUSD).
		Suffix("token_b_usd = ?,", swapCount.TokenBUSD).
		Suffix("tvl_in_usd = ?,", swapCount.TvlInUsd).
		Suffix("tx_num = swap_count_k_lines.tx_num + 1,").
		Suffix("vol_in_usd = swap_count_k_lines.vol_in_usd + ?,", swapCount.VolInUsd).
		Suffix(avgFmt, swapCount.Avg).
		Suffix("WHERE ").
		Suffix("swap_count_k_lines.last_swap_transaction_id <= ?", swapCount.LastSwapTransactionID).
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
			"user_token_a_volume":      userSwapCount.UserTokenAVolume.Abs(),
			"user_token_b_volume":      userSwapCount.UserTokenBVolume.Abs(),
			"user_token_a_balance":     userSwapCount.UserTokenABalance,
			"user_token_b_balance":     userSwapCount.UserTokenBBalance,
			"updated_at":               &now,
			"created_at":               &now,
			"tx_num":                   1,
		}
	)

	sql, args, err := sq.Insert("user_swap_counts").SetMap(inserts).Suffix("ON CONFLICT(user_address,swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", userSwapCount.LastSwapTransactionID).
		Suffix("user_token_a_volume = user_swap_counts.user_token_a_volume + ?,", userSwapCount.UserTokenAVolume.Abs()).
		Suffix("user_token_b_volume = user_swap_counts.user_token_b_volume + ?,", userSwapCount.UserTokenBVolume.Abs()).
		Suffix("user_token_a_balance = ?,", userSwapCount.UserTokenABalance).
		Suffix("user_token_b_balance = ?,", userSwapCount.UserTokenBBalance).
		Suffix("tx_num = user_swap_counts.tx_num +1").
		Suffix("WHERE ").
		Suffix("user_swap_counts.last_swap_transaction_id <= ?", userSwapCount.LastSwapTransactionID).
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
			"user_token_a_volume":      userSwapCount.UserTokenAVolume.Abs(),
			"user_token_b_volume":      userSwapCount.UserTokenBVolume.Abs(),
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
		Suffix("user_token_a_volume = user_swap_count_days.user_token_a_volume + ?,", userSwapCount.UserTokenAVolume.Abs()).
		Suffix("user_token_b_volume = user_swap_count_days.user_token_b_volume + ?,", userSwapCount.UserTokenBVolume.Abs()).
		Suffix("user_token_a_balance = ?,", userSwapCount.UserTokenABalance).
		Suffix("user_token_b_balance = ?,", userSwapCount.UserTokenBBalance).
		Suffix("tx_num = user_swap_count_days.tx_num +1").
		Suffix("WHERE ").
		Suffix("user_swap_count_days.last_swap_transaction_id <= ?", userSwapCount.LastSwapTransactionID).
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

func GetLastSwapCountByGroup(ctx context.Context, filter ...Filter) (*domain.SwapCount, error) {
	var swapCount *domain.SwapCount
	if err := wDB(ctx).Model(&domain.SwapCount{}).Scopes(filter...).Order("id desc").First(&swapCount).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return swapCount, nil
}

func GetLastMaxTvls(ctx context.Context, filter ...Filter) ([]*domain.SwapCount, error) {
	var ids []int64
	if err := wDB(ctx).Model(&domain.SwapCount{}).Scopes(filter...).Select("max(ix)").Group("swap_address").Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	if ids == nil {
		return nil, nil
	}

	var list []*domain.SwapCount
	wDB(ctx).Model(&domain.SwapCount{}).Select("tvl,swap_address").Where("id in ?", ids).Scan(&list)
	return list, nil
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

func CountUserSwapCount(ctx context.Context, filter ...Filter) (int64, error) {
	var (
		db    = rDB(ctx)
		total int64
		err   error
	)
	if err = db.Model(&domain.UserSwapCount{}).Scopes(filter...).Count(&total).Error; err != nil {
		return 0, errors.Wrap(err)
	}

	return total, nil
}

func QueryUserSwapCounts(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.UserSwapCount, error) {
	var (
		db   = rDB(ctx)
		list []*domain.UserSwapCount
		err  error
	)

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return list, nil
}

func QuerySwapCount(ctx context.Context, filter ...Filter) (*domain.SwapCount, error) {
	var swapCount = &domain.SwapCount{}
	if err := rDB(ctx).Model(&domain.SwapCount{}).Scopes(filter...).Take(swapCount).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return swapCount, nil

}

func QuerySwapCountKLine(ctx context.Context, filter ...Filter) (*domain.SwapCountKLine, error) {
	var (
		db             = rDB(ctx)
		err            error
		swapCountKLine = &domain.SwapCountKLine{}
	)

	if err = db.Model(&domain.SwapCountKLine{}).Scopes(filter...).Take(swapCountKLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapCountKLine, nil
}

func QuerySwapCountKLines(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapCountKLine, error) {
	var (
		db             = rDB(ctx)
		err            error
		swapCountKLine []*domain.SwapCountKLine
	)

	if err = db.Model(&domain.SwapCountKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapCountKLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapCountKLine, nil
}

func SumSwapCountVolForKLines(ctx context.Context, filter ...Filter) (*domain.SwapCountKLineVolCount, error) {
	var (
		db                     = rDB(ctx)
		err                    error
		swapCountKLineVolCount = &domain.SwapCountKLineVolCount{}
	)

	if err = db.Model(&domain.SwapCountKLine{}).Select("SUM(token_a_volume) as token_a_volume, SUM(token_b_volume) as token_b_volume, SUM(tx_num) as tx_num").Scopes(filter...).Scan(&swapCountKLineVolCount).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapCountKLineVolCount, nil
}
