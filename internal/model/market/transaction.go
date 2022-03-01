package model

import (
	"context"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/pkg/domain"
)

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
	if err := wDB(ctx).Model(&domain.SwapTransaction{}).Scopes(filter...).Take(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}

func CountTxNum(ctx context.Context, filter ...Filter) (*domain.SumVol, error) {
	var (
		sumTokenAVol *domain.SumVol
		sumTokenBVol *domain.SumVol
	)
	if err := wDB(ctx).Debug().Model(&domain.SwapTransaction{}).Select("count(*) as tx_num , sum(token_a_volume) as token_a_total_vol").Scopes(filter...).Where("token_a_volume < 0").Take(&sumTokenAVol).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	if err := wDB(ctx).Debug().Model(&domain.SwapTransaction{}).Select("count(*) as tx_num , sum(token_b_volume) as token_b_total_vol").Scopes(filter...).Where("token_b_volume < 0").Take(&sumTokenBVol).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &domain.SumVol{
		TxNum:          sumTokenAVol.TxNum + sumTokenBVol.TxNum,
		TokenATotalVol: sumTokenAVol.TokenATotalVol.Abs(),
		TokenBTotalVol: sumTokenBVol.TokenBTotalVol.Abs(),
	}, nil
}
