package model

import (
	"context"
	"time"

	dbPool "git.cplus.link/go/akit/client/psql"
	"git.cplus.link/go/akit/errors"
	sq "github.com/Masterminds/squirrel"
	"gorm.io/gorm"
	"gorm.io/hints"

	"git.cplus.link/crema/backend/pkg/domain"
)

func CreateSwapPairCount(ctx context.Context, tokenVolumeCount *domain.SwapPairCountSharding) error {
	if err := wDB(ctx).Create(tokenVolumeCount).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapPairCount(ctx context.Context, filter ...Filter) (*domain.SwapPairCountSharding, error) {
	var (
		db   = rDB(ctx)
		info *domain.SwapPairCountSharding
	)
	if err := db.Model(&domain.SwapPairCountSharding{}).Scopes(filter...).Order("id desc").First(&info).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return info, nil
}

func QueryTvl(ctx context.Context, filter ...Filter) (*domain.Tvl, error) {
	var (
		db   = rDB(ctx)
		info *domain.Tvl
	)
	if err := db.Model(&domain.Tvl{}).Scopes(filter...).Order("id desc").First(&info).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return info, nil
}

func CreateTvl(ctx context.Context, tvl *domain.Tvl) error {
	if err := wDB(ctx).Create(tvl).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapPairPriceKLine(ctx context.Context, filter ...Filter) (*domain.SwapPairPriceKLine, error) {
	var (
		db                 = rDB(ctx)
		err                error
		swapPairPriceKLine = &domain.SwapPairPriceKLine{}
	)

	if err = db.Model(&domain.SwapPairPriceKLine{}).Scopes(filter...).Take(swapPairPriceKLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapPairPriceKLine, nil
}

func QuerySwapPairPriceKLines(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapPairPriceKLine, error) {
	var (
		db                  = rDB(ctx)
		err                 error
		swapPairPriceKLines []*domain.SwapPairPriceKLine
	)

	if err = db.Model(&domain.SwapPairPriceKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapPairPriceKLines).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapPairPriceKLines, nil
}

func QuerySwapPairPriceKLinesInBaseTable(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapPairPriceKLine, error) {
	var (
		db                  = rDB(ctx)
		err                 error
		swapPairPriceKLines []*domain.SwapPairPriceKLine
	)

	if err = db.Clauses(hints.Comment("select", "nosharding")).Model(&domain.SwapPairPriceKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapPairPriceKLines).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapPairPriceKLines, nil
}

func CreateSwapPairPriceKLine(ctx context.Context, transactions []*domain.SwapPairPriceKLine) error {
	if err := wDB(ctx).Create(transactions).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpsertSwapPairPriceKLine(ctx context.Context, swapPairPriceKLine *domain.SwapPairPriceKLine) (*domain.SwapPairPriceKLine, error) {
	var (
		after   domain.SwapPairPriceKLine
		now     = time.Now().UTC()
		avgFmt  string
		inserts = map[string]interface{}{
			"swap_address": swapPairPriceKLine.SwapAddress,
			"date_type":    swapPairPriceKLine.DateType,
			"open":         swapPairPriceKLine.Open,
			"high":         swapPairPriceKLine.High,
			"low":          swapPairPriceKLine.Low,
			"settle":       swapPairPriceKLine.Settle,
			"avg":          swapPairPriceKLine.Avg,
			"num":          1,
			"updated_at":   &now,
			"created_at":   &now,
			"date":         swapPairPriceKLine.Date,
		}
	)
	// fullName := "swap_pair_price_k_lines" // todo 取消
	fullName, err := getTableFullName(after.TableName(), swapPairPriceKLine.SwapAddress)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 除了domain.DateMin 类型，其他的都是根据前一个类型求平均值
	if swapPairPriceKLine.DateType == domain.DateMin {
		avgFmt = "avg = (" + fullName + ".avg * " + fullName + ".num + ? )/(" + fullName + ".num+ 1)"
	} else {
		avgFmt = "avg = ?"
	}

	sql, args, err := sq.Insert("swap_pair_price_k_lines").SetMap(inserts).Suffix("ON CONFLICT(swap_address,date,date_type) DO UPDATE SET").
		Suffix("high = ?,", swapPairPriceKLine.High).
		Suffix("low = ?,", swapPairPriceKLine.Low).
		Suffix("settle = ?,", swapPairPriceKLine.Settle).
		Suffix("num = "+fullName+".num + 1,").
		Suffix(avgFmt, swapPairPriceKLine.Avg).
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

func DeleteSwapPairPriceKLine(ctx context.Context, filter ...Filter) error {
	res := wDB(ctx).Scopes(filter...).Delete(&domain.SwapPairPriceKLine{})
	if err := res.Error; err != nil {
		return errors.Wrap(err)
	}
	if res.RowsAffected <= 0 {
		return errors.Wrap(errors.RecordNotFound)
	}
	return nil
}

func QuerySwapTokenPriceKLine(ctx context.Context, filter ...Filter) (*domain.SwapTokenPriceKLine, error) {
	var (
		db                  = rDB(ctx)
		err                 error
		swapTokenPriceKLine = &domain.SwapTokenPriceKLine{}
	)

	if err = db.Model(&domain.SwapTokenPriceKLine{}).Scopes(filter...).Take(swapTokenPriceKLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return swapTokenPriceKLine, errors.Wrap(errors.RecordNotFound)
		}
		return swapTokenPriceKLine, errors.Wrap(err)
	}

	return swapTokenPriceKLine, nil
}

func QuerySwapTokenPriceKLines(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTokenPriceKLine, error) {
	var (
		db                   = rDB(ctx)
		err                  error
		swapTokenPriceKLines []*domain.SwapTokenPriceKLine
	)

	if err = db.Model(&domain.SwapTokenPriceKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapTokenPriceKLines).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapTokenPriceKLines, nil
}

func UpsertSwapTokenPriceKLine(ctx context.Context, swapTokenPriceKLine *domain.SwapTokenPriceKLine) (*domain.SwapTokenPriceKLine, error) {
	var (
		after   domain.SwapTokenPriceKLine
		now     = time.Now().UTC()
		avgFmt  string
		inserts = map[string]interface{}{
			"symbol":     swapTokenPriceKLine.Symbol,
			"date_type":  swapTokenPriceKLine.DateType,
			"open":       swapTokenPriceKLine.Open,
			"high":       swapTokenPriceKLine.High,
			"low":        swapTokenPriceKLine.Low,
			"settle":     swapTokenPriceKLine.Settle,
			"avg":        swapTokenPriceKLine.Avg,
			"num":        1,
			"updated_at": &now,
			"created_at": &now,
			"date":       swapTokenPriceKLine.Date,
		}
	)

	// 除了domain.DateMin 类型，其他的都是根据前一个类型求平均值
	if swapTokenPriceKLine.DateType == domain.DateMin {
		avgFmt = "avg = (swap_token_price_k_lines.avg * swap_token_price_k_lines.num + ? )/(swap_token_price_k_lines.num+ 1)"
	} else {
		avgFmt = "avg = ?"
	}

	sql, args, err := sq.Insert("swap_token_price_k_lines").SetMap(inserts).Suffix("ON CONFLICT(symbol,date,date_type) DO UPDATE SET").
		Suffix("high = ?,", swapTokenPriceKLine.High).
		Suffix("low = ?,", swapTokenPriceKLine.Low).
		Suffix("settle = ?,", swapTokenPriceKLine.Settle).
		Suffix("num = swap_token_price_k_lines.num + 1,").
		Suffix(avgFmt, swapTokenPriceKLine.Avg).
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

func DeleteSwapTokenPriceKLine(ctx context.Context, filter ...Filter) error {
	res := wDB(ctx).Scopes(filter...).Delete(&domain.SwapTokenPriceKLine{})
	if err := res.Error; err != nil {
		return errors.Wrap(err)
	}

	if res.RowsAffected <= 0 {
		return errors.Wrap(errors.RecordNotFound)
	}
	return nil
}

func CountUserNumber(ctx context.Context) (int64, error) {
	var (
		db    = rDB(ctx)
		err   error
		total int64
	)

	if err = db.Model(&domain.UserCountKLine{}).Select("select count(*) from user_count_k_lines where date_type = 'mon' ").Group("user_address").Count(&total).Error; err != nil {
		return 0, errors.Wrap(err)
	}
	return total, nil
}

func GetSwapTransactionV2(ctx context.Context, filter ...Filter) (*domain.SwapTransactionV2, error) {
	var (
		info domain.SwapTransactionV2
	)
	if err := rDB(ctx).Scopes(filter...).First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.RecordNotFound
		}
		return nil, errors.Wrap(err)
	}
	return &info, nil
}

func CreatedSwapTransactionV2(ctx context.Context, transactionV2 *domain.SwapTransactionV2) error {
	if err := wDB(ctx).Create(transactionV2).Error; err != nil {
		if dbPool.IsDuplicateKeyError(err) {
			return errors.Wrap(errors.AlreadyExists)
		}
		return errors.Wrap(err)
	}
	return nil
}
