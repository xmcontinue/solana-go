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
