package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

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
