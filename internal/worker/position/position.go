package position

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
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
	logger.Info("rand sync time", logger.String("rand time：", before.String()))
	if err != nil {
		logger.Info("rand sync time", logger.Errorv(err))
		return nil
	}

	// 1.
	// 查询当天是否已同步
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
		if swapPair.Version == "v2" {
			continue
		}
		// 3.1 获取swap池子仓位
		swapAccountAndPositionsAccount, err := sol.GetSwapAccountAndPositionsAccountForSwapKey(swapPair.SwapPublicKey)
		if err != nil {
			return errors.Wrap(err)
		}

		// 3.3 解析至model
		positionsMode, err = positionsAccountToModel(swapPair, tokenPrices, positionsMode, swapAccountAndPositionsAccount)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	// 4.获取farming池子仓位
	farmingPositions, err := GetFarmingPositions(solana.MustPublicKeyFromBase58("CPWdCBwzgC2MNQKz7AGAkZH51BskgA1LY9v8RPikQ2x1"))
	if err != nil {
		return errors.Wrap(err)
	}

	// 5.替换掉仓位用户地址为farming仓位地址
	farmingPositionsM := make(map[string]sol.FarmingPosition, len(farmingPositions))
	for _, v := range farmingPositions {
		farmingPositionsM[v.NftMint.String()] = v
	}

	for _, v := range positionsMode {
		if p, has := farmingPositionsM[v.PositionID]; has {
			v.UserAddress = p.Owner.String()
		}
	}

	// 6.写入db
	err = model.CreateSwapPositionCountSnapshots(context.Background(), positionsMode)
	if err != nil {
		return errors.Wrap(err)
	}

	logger.Info("sync success")
	return nil
}

func positionsAccountToModel(swapPair *domain.SwapConfig, tokenPrices map[string]decimal.Decimal, positionsMode []*domain.PositionCountSnapshot, swapAccountAndPositionsAccount *sol.SwapAccountAndPositionsAccount) ([]*domain.PositionCountSnapshot, error) {
	for _, positionAccount := range swapAccountAndPositionsAccount.PositionsAccount {
		for k, v := range positionAccount.Positions {
			// 通过tokenID获取user address
			logger.Info(fmt.Sprintf("sync user address: swap address(%s) ,total num(%d), now num(%d) ", swapAccountAndPositionsAccount.SwapAccount.TokenSwapKey.String(), len(positionAccount.Positions), k))
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
				SwapAddress:  swapAccountAndPositionsAccount.SwapAccount.TokenSwapKey.String(),
				PositionID:   v.NftTokenId.String(),
				Date:         before.Format("2006-01-02 15:04:05"),
				TokenAAmount: parse.PrecisionConversion(tokenAAmount, int(swapPair.TokenA.Decimal)),
				TokenBAmount: parse.PrecisionConversion(tokenBAmount, int(swapPair.TokenB.Decimal)),
				TokenAPrice:  tokenAPrice,
				TokenBPrice:  tokenBPrice,
				Raw:          positionAccount.PositionsRaw[k],
			})
		}
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

func GetFarmingPositions(positionWrapper solana.PublicKey) ([]sol.FarmingPosition, error) {
	b, _ := base58.Decode("Ahg1opVcGX")
	opt := &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{
			{
				DataSize: 306,
			},
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Offset: 177,
					Bytes:  b,
				},
			},
		},
	}
	res, err := sol.GetRpcClient().GetProgramAccountsWithOpts(context.Background(), positionWrapper, opt)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	positions := make([]sol.FarmingPosition, 0)
	for _, v := range res {
		position := sol.FarmingPosition{}
		err = bin.NewBinDecoder(v.Account.Data.GetBinary()[8:]).Decode(&position)
		positions = append(positions, position)
	}
	return positions, nil
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

func syncPositionV2() error {
	logger.Info("sync position V2 start")

	var err error
	//err = randTime()
	logger.Info("rand sync time", logger.String("rand time：", before.String()))
	if err != nil {
		logger.Info("rand sync time", logger.Errorv(err))
		return nil
	}

	_, err = model.QuerySwapPositionV2Snapshot(context.Background(), model.NewFilter("date > ?", timeZero(time.Now())))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Info("sync skipped today")
		return nil
	}

	swapList := sol.SwapConfigList()

	domainPositionV2List := make([]*domain.PositionV2Snapshot, 0, len(swapList))
	for _, swap := range swapList {
		if swap.Version != "v2" {
			continue
		}

		positionV2, err := sol.GetPositionInfoV2(swap.SwapPublicKey)
		if err != nil {
			logger.Error("GetPositionInfoV2 err", logger.Errorv(err))
			return errors.Wrap(err)
		}
		domainPositionV2 := positionV2ToModel(positionV2)
		domainPositionV2List = append(domainPositionV2List, domainPositionV2...)

	}

	err = model.CreateSwapPositionV2Snapshot(context.Background(), domainPositionV2List)
	if err != nil {
		logger.Error("CreateSwapPositionV2Snapshot ", logger.Errorv(err))
		return errors.Wrap(err)
	}
	return nil
}

func positionV2ToModel(positionV2s []*sol.PositionV2) []*domain.PositionV2Snapshot {
	domainPositionV2 := make([]*domain.PositionV2Snapshot, 0, len(positionV2s))
	for i := range positionV2s {
		domainPositionV2 = append(domainPositionV2, &domain.PositionV2Snapshot{
			Model:            gorm.Model{},
			ClmmPool:         positionV2s[i].ClmmPool.String(),
			PositionNFTMint:  positionV2s[i].PositionNFTMint.String(),
			Date:             before.Format("2006-01-02 15:04:05"),
			Liquidity:        positionV2s[i].Liquidity.Val(),
			TickLowerIndex:   positionV2s[i].TickLowerIndex,
			TickUpperIndex:   positionV2s[i].TickUpperIndex,
			FeeGrowthInsideA: positionV2s[i].FeeGrowthInsideA.Val(),
			FeeOwedA:         positionV2s[i].FeeOwedA,
			FeeGrowthInsideB: positionV2s[i].FeeGrowthInsideB.Val(),
			FeeOwedB:         positionV2s[i].FeeOwedB,
			//GrowthInside1:    positionV2s[i].RewardInfos[0].GrowthInside.Val(),
			//AmountOwed1:      positionV2s[i].RewardInfos[0].AmountOwed,
			//GrowthInside2:    positionV2s[i].RewardInfos[1].GrowthInside.Val(),
			//AmountOwed2:      positionV2s[i].RewardInfos[1].AmountOwed,
			//GrowthInside3:    positionV2s[i].RewardInfos[2].GrowthInside.Val(),
			//AmountOwed3:      positionV2s[i].RewardInfos[2].AmountOwed,
		})
	}

	return domainPositionV2
}
