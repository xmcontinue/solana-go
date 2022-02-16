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
