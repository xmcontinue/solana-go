package position

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func syncPosition() error {
	// 1.查询当天是否已同步
	_, err := model.QuerySwapPositionCountSnapshot(context.Background(), model.NewFilter("created_at = ?", time.Now().Format("2006-01-02")))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	// 2.获取币价
	swapList := sol.SwapConfigList()
	tokenPrices, err := getTokenPriceForPairList(swapList)
	if err != nil {
		return errors.Wrap(err)
	}

	// 2.同步仓位数据
	positionsMode := make([]*domain.PositionCountSnapshot, 0)
	for _, swapPair := range swapList {
		// 2.1 获取池子仓位
		positionsAccount, err := sol.GetPositionsAccountForSwapKey(swapPair.SwapPublicKey)
		if err != nil {
			return errors.Wrap(err)
		}

		// 2.2 解析至model
		err = positionsAccountToModel(swapPair, tokenPrices, positionsMode, positionsAccount)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	// 4.写入db
	err = model.CreateSwapPositionCountSnapshots(context.Background(), positionsMode)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func positionsAccountToModel(swapPair *domain.SwapConfig, tokenPrices map[string]decimal.Decimal, positionsMode []*domain.PositionCountSnapshot, positionsAccount *sol.PositionsAccount) error {
	for k, v := range positionsAccount.Positions {
		// 2.3 通过tokenID获取user address
		userAddress, err := sol.GetUserAddressForTokenKey(v.NftTokenId)
		if err != nil {
			return errors.Wrap(err)
		}
		// 2.5 计算
		tokenAPrice, tokenBPrice := tokenPrices[swapPair.TokenA.Symbol], tokenPrices[swapPair.TokenB.Symbol]
		// todo
		TokenAAmount, TokenBAmount := decimal.Decimal{}, decimal.Decimal{}

		positionsMode = append(positionsMode, &domain.PositionCountSnapshot{
			UserAddress:  userAddress,
			SwapAddress:  positionsAccount.TokenSwapKey.String(),
			Date:         time.Now().Format("2006-01-02"),
			TokenAAmount: TokenAAmount,
			TokenBAmount: TokenBAmount,
			TokenAPrice:  tokenAPrice,
			TokenBPrice:  tokenBPrice,
			Raw:          positionsAccount.PositionsRaw[k],
		})
	}

	return nil
}

func getTokenPriceForPairList(swapList []*domain.SwapConfig) (map[string]decimal.Decimal, error) {
	tokenPrices := make(map[string]decimal.Decimal, 0)
	for _, v := range swapList {

		if _, ok := tokenPrices[v.TokenA.Symbol]; !ok {
			tokenAPrice, err := model.GetPriceForSymbol(context.Background(), v.TokenA.Symbol)
			if err != nil {
				return nil, errors.Wrap(err)
			}
			tokenPrices[v.TokenA.Symbol] = tokenAPrice
		}

		if _, ok := tokenPrices[v.TokenB.Symbol]; !ok {
			tokenBPrice, err := model.GetPriceForSymbol(context.Background(), v.TokenB.Symbol)
			if err != nil {
				return nil, errors.Wrap(err)
			}
			tokenPrices[v.TokenB.Symbol] = tokenBPrice
		}
	}
	return tokenPrices, nil
}
