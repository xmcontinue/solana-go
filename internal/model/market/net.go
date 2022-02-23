package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QueryNetRecords(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.NetRecode, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.NetRecode
		total int64
		err   error
	)
	if err = db.Model(&domain.NetRecode{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Order("id asc").Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil
}
