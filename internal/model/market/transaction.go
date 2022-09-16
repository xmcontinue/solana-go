package model

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QuerySwapPairBase(ctx context.Context, filter ...Filter) (*domain.SwapPairBase, error) {
	var info *domain.SwapPairBase
	if err := wDB(ctx).Model(&domain.SwapPairBase{}).Scopes(filter...).First(&info).Error; err != nil {
		return info, errors.Wrap(err)
	}
	return info, nil
}

func QuerySwapPairBases(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapPairBase, error) {
	var (
		db        = rDB(ctx)
		err       error
		pairBases []*domain.SwapPairBase
	)

	if err = db.Model(&domain.SwapPairBase{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&pairBases).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return pairBases, nil
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

func UpdateSwapTransaction(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	db := wDB(ctx).Model(&domain.SwapTransaction{}).Scopes(filter...).Updates(updates)
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

func GetPriceForSymbol(ctx context.Context, symbol string, filter ...Filter) (decimal.Decimal, error) {
	var (
		info  *domain.SwapTokenPriceKLine
		price = decimal.NewFromInt(0)
	)
	newFilters := make([]Filter, 0, len(filter)+3)
	newFilters = append(newFilters, NewFilter("date_type = ?", "1min"))
	newFilters = append(newFilters, NewFilter("symbol = ?", symbol))
	if len(filter) != 0 {
		newFilters = append(newFilters, filter...)
	}
	newFilters = append(newFilters, OrderFilter("date desc"))
	if err := wDB(ctx).Model(&domain.SwapTokenPriceKLine{}).Scopes(newFilters...).Take(&info).Error; err != nil {
		return price, errors.Wrap(err)
	}

	return info.Settle, nil
}
