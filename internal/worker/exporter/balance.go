package exporter

import (
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
	"git.cplus.link/crema/backend/pkg/prometheus"
)

func WatchBalance() error {
	swapPairs := sol.SwapConfigList()

	positionsMode := make([]*domain.PositionCountSnapshot, 0)
	for _, pair := range swapPairs {
		// TODO 暂时只监听一个swap池子
		if pair.SwapAccount != "EzvTazMwHjLAg3oVKK7LCBFP3ThEqkLDiW55frQbQTby" {
			continue
		}

		// 获取swap池子仓位
		swapAccountAndPositionsAccount, err := sol.GetSwapAccountAndPositionsAccountForSwapKey(pair.SwapPublicKey)
		if err != nil {
			continue
		}

		// 解析至model
		positionsMode, err = positionsAccountToModel(pair, positionsMode, swapAccountAndPositionsAccount)
		if err != nil {
			return errors.Wrap(err)
		}

		totalTokenAAmount, totalTokenBAmount := decimal.Decimal{}, decimal.Decimal{}
		for _, position := range positionsMode {
			totalTokenAAmount = totalTokenAAmount.Add(position.TokenAAmount)
			totalTokenBAmount = totalTokenBAmount.Add(position.TokenBAmount)
		}

		TokenADifference := totalTokenAAmount.Sub(pair.TokenA.Balance)
		TokenBDifference := totalTokenBAmount.Sub(pair.TokenB.Balance)

		if pair.TokenA.Balance.IsZero() || pair.TokenB.Balance.IsZero() {
			sendBalanceRateMsgToPushGateway(1, pair.SwapAccount, "swap pool balance is zero")
			continue
		}

		tokenARate := TokenADifference.Div(totalTokenAAmount).Round(6)
		if tokenARate.Cmp(decimal.NewFromFloat(0.02)) > 0 || tokenARate.Cmp(decimal.NewFromFloat(-0.02)) < 0 {
			f, _ := tokenARate.Float64()
			sendBalanceRateMsgToPushGateway(f, pair.SwapAccount, "too much difference")
		}
		sendBalanceRateMsgToPushGateway(0.03, pair.SwapAccount, "too much difference")
		tokenBRate := TokenBDifference.Div(totalTokenBAmount).Round(6)
		if tokenBRate.Cmp(decimal.NewFromFloat(0.02)) > 0 || tokenBRate.Cmp(decimal.NewFromFloat(-0.02)) < 0 {
			f, _ := tokenBRate.Float64()
			sendBalanceRateMsgToPushGateway(f, pair.SwapAccount, "too much difference")
		}
	}
	return nil
}

func positionsAccountToModel(swapPair *domain.SwapConfig, positionsMode []*domain.PositionCountSnapshot, swapAccountAndPositionsAccount *sol.SwapAccountAndPositionsAccount) ([]*domain.PositionCountSnapshot, error) {
	for _, v := range swapAccountAndPositionsAccount.Positions {
		// 通过tokenID获取user address
		userAddress, err := sol.GetUserAddressForTokenKey(v.NftTokenId)
		if err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				continue
			}
			return nil, errors.Wrap(err)
		}
		// 计算 amount
		tokenAAmount, tokenBAmount := swapAccountAndPositionsAccount.CalculateTokenAmount(&v)
		positionsMode = append(positionsMode, &domain.PositionCountSnapshot{
			UserAddress:  userAddress,
			SwapAddress:  swapAccountAndPositionsAccount.TokenSwapKey.String(),
			PositionID:   v.NftTokenId.String(),
			TokenAAmount: parse.PrecisionConversion(tokenAAmount, int(swapPair.TokenA.Decimal)),
			TokenBAmount: parse.PrecisionConversion(tokenBAmount, int(swapPair.TokenB.Decimal)),
		})
	}

	return positionsMode, nil
}

func sendBalanceRateMsgToPushGateway(value float64, swapAddress string, msg string) {
	log := &iface.LogReq{
		LogName:  "balance_rate",
		LogValue: value,
		LogHelp:  "Comparison of current balance with current liquidity",
		JobName:  "balance_rate",
		Tags: map[string]string{
			"project":  prometheus.GetProjectName(),
			"swap_key": swapAddress,
			"msg":      msg,
		},
	}
	err := prometheus.ExamplePusherPush(log)
	if err != nil {
		logger.Error("send msg to push_gateway failed!", logger.Errorv(err))
	}
}
