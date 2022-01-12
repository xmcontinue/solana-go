package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/common/pkg/domain"
)

func CreateSwapPairCount(ctx context.Context, tokenVolumeCount *domain.SwapPairCount) error {
	if err := wDB(ctx).Create(tokenVolumeCount).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapPairCount(ctx context.Context, filter ...Filter) (*domain.SwapPairCount, error) {
	var (
		db   = rDB(ctx)
		info *domain.SwapPairCount
	)
	if err := db.Model(&domain.SwapPairCount{}).Scopes(filter...).Order("id desc").First(&info).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return info, nil
}
