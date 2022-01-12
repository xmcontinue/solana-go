package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/common/pkg/domain"
)

func CreateTokenVolumeCount(ctx context.Context, tokenVolumeCount *domain.TokenVolumeCount) error {
	if err := wDB(ctx).Create(tokenVolumeCount).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QueryTokenVolumeCount(ctx context.Context, filter ...Filter) (*domain.TokenVolumeCount, error) {
	var (
		db   = rDB(ctx)
		info *domain.TokenVolumeCount
	)
	if err := db.Model(&domain.TokenVolumeCount{}).Scopes(filter...).Order("id desc").First(&info).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return info, nil
}
