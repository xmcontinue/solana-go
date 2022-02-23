package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/pkg/domain"
)

func CreateBaseTransactions(ctx context.Context, transactions []*domain.TransactionBase) error {
	if err := wDB(ctx).Create(transactions).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QueryBaseTransaction(ctx context.Context, filter ...Filter) (*domain.TransactionBase, error) {
	var transaction *domain.TransactionBase
	if err := wDB(ctx).Model(&domain.TransactionBase{}).Scopes(filter...).Order("id desc").First(&transaction).Error; err != nil {
		return transaction, errors.Wrap(err)
	}
	return transaction, nil
}

func QuerySwapPairBase(ctx context.Context, filter ...Filter) (*domain.SwapPairBase, error) {
	var info *domain.SwapPairBase
	if err := wDB(ctx).Model(&domain.SwapPairBase{}).Scopes(filter...).First(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}

func CreateSwapPairBase(ctx context.Context, swapPairBase *domain.SwapPairBase) error {
	if err := wDB(ctx).Create(swapPairBase).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpdateSwapPairBase(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	db := wDB(ctx).Model(&domain.SwapPairBase{}).Scopes(filter...).Updates(updates)
	if err := db.Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func CreateSwapTransactions(ctx context.Context, transactions []*domain.SwapTransaction) error {
	if err := wDB(ctx).Create(transactions).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapTransaction(ctx context.Context, filter ...Filter) (*domain.SwapTransaction, error) {
	var info *domain.SwapTransaction
	if err := wDB(ctx).Model(&domain.SwapTransaction{}).Scopes(filter...).First(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}
