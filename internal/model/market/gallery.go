package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/pkg/domain"
)

func CreateMetadataJson(ctx context.Context, metadataJson *domain.MetadataJsonDate) error {
	if err := wDB(ctx).Create(metadataJson).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QueryMetadataJson(ctx context.Context, filter ...Filter) (*domain.MetadataJsonDate, error) {
	var info *domain.MetadataJsonDate
	if err := wDB(ctx).Model(&domain.MetadataJsonDate{}).Scopes(filter...).Take(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}
