package model

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QuerySwapTransactionsV2(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTransactionV2, error) {
	var (
		db   = rDB(ctx)
		list []*domain.SwapTransactionV2

		err error
	)

	if err = db.Model(&domain.SwapTransactionV2{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	if len(list) == 0 {
		return nil, errors.RecordNotFound
	}
	return list, nil
}

func DeleteSwapTransactionV2(ctx context.Context, filter ...Filter) error {
	res := wDB(ctx).Scopes(filter...).Delete(&domain.SwapTransactionV2{})
	if err := res.Error; err != nil {
		return errors.Wrap(err)
	}
	if res.RowsAffected <= 0 {
		return errors.Wrap(errors.RecordNotFound)
	}
	return nil
}

func QueryTransActionUserCount(ctx context.Context, filter ...Filter) (*domain.TransActionUserCount, error) {
	var (
		info domain.TransActionUserCount
	)
	if err := rDB(ctx).Scopes(filter...).Take(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.RecordNotFound
		}
		return nil, errors.Wrap(err)
	}
	return &info, nil
}

func CountTransActionUserCount(ctx context.Context) (int64, error) {
	var (
		db    = rDB(ctx)
		err   error
		total int64
	)

	if err = db.Model(&domain.TransActionUserCount{}).Count(&total).Error; err != nil {
		return 0, errors.Wrap(err)
	}
	return total, nil
}

func CreateTransActionUserCount(ctx context.Context, info *domain.TransActionUserCount) error {
	if err := wDB(ctx).Create(info).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapUserCount(ctx context.Context, filter ...Filter) (*domain.SwapUserCount, error) {
	var (
		info domain.SwapUserCount
	)
	if err := rDB(ctx).Scopes(filter...).Take(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.RecordNotFound
		}
		return nil, errors.Wrap(err)
	}
	return &info, nil
}

func CreateSwapUserCount(ctx context.Context, info *domain.SwapUserCount) error {
	if err := wDB(ctx).Create(info).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpdateSwapUserCount(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	db := wDB(ctx).Model(&domain.SwapUserCount{}).Scopes(filter...).Updates(updates)
	if err := db.Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}
