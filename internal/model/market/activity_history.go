package model

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/pkg/domain"
)

func CreateActivityHistory(ctx context.Context, history *domain.ActivityHistory) error {
	if err := wDB(ctx).Create(history).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpdateActivityHistory(ctx context.Context, id int64, updates map[string]interface{}) error {
	var (
		db = wDB(ctx)
	)
	if err := db.Model(&domain.ActivityHistory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func SelectByUser(ctx context.Context, user string) ([]*domain.ActivityHistory, error) {
	var (
		db   = rDB(ctx)
		list []*domain.ActivityHistory
	)
	if err := db.Where("user_key = ?", user).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return list, nil
}

func SelectLatest(ctx context.Context) (*domain.ActivityHistory, error) {
	var (
		db  = rDB(ctx)
		res *domain.ActivityHistory
	)
	if err := db.Order("block_time desc").First(&res).Error; err == nil {
		return res, nil
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else {
		return nil, errors.Wrap(err)
	}

}

func SelectByUserMint(ctx context.Context, user, mint string) (*domain.ActivityHistory, error) {
	var (
		db  = rDB(ctx)
		res *domain.ActivityHistory
	)
	if err := db.Where("user_key = ? and mint_key = ?", user, mint).First(&res).Error; err == nil {
		return res, nil
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else {
		return nil, errors.Wrap(err)
	}
}
