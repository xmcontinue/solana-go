package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/pkg/domain"
)

func CreateSwapPositionCountSnapshots(ctx context.Context, positionCountSnapshots []*domain.PositionCountSnapshot) error {
	if err := wDB(ctx).Create(positionCountSnapshots).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapPositionCountSnapshot(ctx context.Context, filter ...Filter) (*domain.PositionCountSnapshot, error) {
	var info *domain.PositionCountSnapshot
	if err := wDB(ctx).Model(&domain.PositionCountSnapshot{}).Scopes(filter...).Take(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}

func CreateSwapPositionV2Snapshot(ctx context.Context, positionV2Snapshots []*domain.PositionV2Snapshot) error {
	if err := wDB(ctx).Create(positionV2Snapshots).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapPositionV2Snapshot(ctx context.Context, filter ...Filter) (*domain.PositionV2Snapshot, error) {
	var info *domain.PositionV2Snapshot
	if err := wDB(ctx).Model(&domain.PositionV2Snapshot{}).Scopes(filter...).Take(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}
