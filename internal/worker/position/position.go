package position

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/rand"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	before = time.Now().Add(-time.Hour * 24)
)

func syncPosition() error {
	logger.Info("sync start")

	err := randTime()
	if err != nil {
		return nil
	}

	logger.Info("rand sync time", logger.String("rand time：", before.String()))
	// 1.查询当天是否已同步
	_, err = model.QuerySwapPositionCountSnapshot(context.Background(), model.NewFilter("date > ?", timeZero(time.Now())))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Info("sync skipped today")
		return nil
	}

	// 2.获取币价
	swapList := sol.SwapConfigList()
	tokenPrices, err := getTokenPriceForPairList(swapList)
	if err != nil {
		return errors.Wrap(err)
	}

	// 3.同步仓位数据
	positionsMode := make([]*domain.PositionCountSnapshot, 0)
	for _, swapPair := range swapList {
		// 3.1 获取池子仓位
		swapAccountAndPositionsAccount, err := sol.GetSwapAccountAndPositionsAccountForSwapKey(swapPair.SwapPublicKey)
		if err != nil {
			return errors.Wrap(err)
		}

		// 3.2 解析至model
		positionsMode, err = positionsAccountToModel(swapPair, tokenPrices, positionsMode, swapAccountAndPositionsAccount)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	// 4.写入db
	err = model.CreateSwapPositionCountSnapshots(context.Background(), positionsMode)
	if err != nil {
		return errors.Wrap(err)
	}

	logger.Info("sync success")
	return nil
}

func positionsAccountToModel(swapPair *domain.SwapConfig, tokenPrices map[string]decimal.Decimal, positionsMode []*domain.PositionCountSnapshot, swapAccountAndPositionsAccount *sol.SwapAccountAndPositionsAccount) ([]*domain.PositionCountSnapshot, error) {
	for k, v := range swapAccountAndPositionsAccount.Positions {
		// 通过tokenID获取user address
		logger.Info(fmt.Sprintf("sync user address: swap address(%s) ,total num(%d), now num(%d) ", swapAccountAndPositionsAccount.SwapAccount.TokenSwapKey.String(), len(swapAccountAndPositionsAccount.Positions), k))
		userAddress, err := sol.GetUserAddressForTokenKey(v.NftTokenId)
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			return nil, errors.Wrap(err)
		}
		// 计算 amount
		tokenAPrice, tokenBPrice := tokenPrices[swapPair.TokenA.Symbol], tokenPrices[swapPair.TokenB.Symbol]

		tokenAAmount, tokenBAmount := swapAccountAndPositionsAccount.CalculateTokenAmount(&v)
		positionsMode = append(positionsMode, &domain.PositionCountSnapshot{
			UserAddress:  userAddress,
			SwapAddress:  swapAccountAndPositionsAccount.TokenSwapKey.String(),
			PositionID:   v.NftTokenId.String(),
			Date:         before.Format("2006-01-02 15:04:05"),
			TokenAAmount: parse.PrecisionConversion(tokenAAmount, int(swapPair.TokenA.Decimal)),
			TokenBAmount: parse.PrecisionConversion(tokenBAmount, int(swapPair.TokenB.Decimal)),
			TokenAPrice:  tokenAPrice,
			TokenBPrice:  tokenBPrice,
			Raw:          swapAccountAndPositionsAccount.PositionsRaw[k],
		})
	}

	return positionsMode, nil
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

func randTime() error {
	now := time.Now()
	zero := timeZero(now)
	if before.Before(zero) {
		r := rand.IntnRange(3600*17, 3600*24-1)
		before = zero.Add(time.Duration(r) * time.Second)
	}

	if now.Before(before) {
		return errors.New("the time has not come")
	}

	return nil
}

func timeZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
