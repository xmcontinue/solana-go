package exporter

import (
	"fmt"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
	"git.cplus.link/crema/backend/pkg/prometheus"
)

type SmsConfig struct {
	AccessKeyId  string
	AccessSecret string
	SignName     string
	PhoneNumbers string
}

var smsConfig = SmsConfig{}

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

		labels := map[string]string{
			"pool":           pair.Name,
			"currentAmountA": pair.TokenA.Balance.Round(6).String(),
			"currentAmountB": pair.TokenB.Balance.Round(6).String(),
			"needAmountA":    totalTokenAAmount.Round(6).String(),
			"needAmountB":    totalTokenBAmount.Round(6).String(),
		}
		
		if pair.TokenA.Balance.IsZero() || pair.TokenB.Balance.IsZero() ||
			totalTokenAAmount.IsZero() || totalTokenBAmount.IsZero() {
			sendBalanceRateMsgToPushGateway(1, pair.SwapAccount, labels)
			continue
		}

		tokenARate := TokenADifference.Div(totalTokenAAmount).Round(6)
		tokenBRate := TokenBDifference.Div(totalTokenBAmount).Round(6)
		fA, _ := tokenARate.Float64()
		fB, _ := tokenBRate.Float64()
		if tokenARate.Cmp(decimal.NewFromFloat(0.02)) > 0 || tokenARate.Cmp(decimal.NewFromFloat(-0.02)) < 0 {
			sendBalanceRateMsgToPushGateway(fA, pair.SwapAccount, labels)
			return nil
		}
		if tokenBRate.Cmp(decimal.NewFromFloat(0.02)) > 0 || tokenBRate.Cmp(decimal.NewFromFloat(-0.02)) < 0 {
			sendBalanceRateMsgToPushGateway(fB, pair.SwapAccount, labels)
			return nil
		}

		sendBalanceRateMsgToPushGateway(fA, pair.SwapAccount, map[string]string{})
	}
	return nil
}

func positionsAccountToModel(swapPair *domain.SwapConfig, positionsMode []*domain.PositionCountSnapshot, swapAccountAndPositionsAccount *sol.SwapAccountAndPositionsAccount) ([]*domain.PositionCountSnapshot, error) {
	for _, positionsAccount := range swapAccountAndPositionsAccount.PositionsAccount {
		for _, v := range positionsAccount.Positions {
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
				SwapAddress:  swapAccountAndPositionsAccount.SwapAccount.TokenSwapKey.String(),
				PositionID:   v.NftTokenId.String(),
				TokenAAmount: parse.PrecisionConversion(tokenAAmount, int(swapPair.TokenA.Decimal)),
				TokenBAmount: parse.PrecisionConversion(tokenBAmount, int(swapPair.TokenB.Decimal)),
			})
		}
	}

	return positionsMode, nil
}

func sendBalanceRateMsgToPushGateway(value float64, swapAddress string, labels map[string]string) {
	log := &iface.LogReq{
		LogName:  "balance_rate",
		LogValue: value,
		LogHelp:  "Comparison of current balance with current liquidity",
		JobName:  "balance_rate",
		Tags: map[string]string{
			"project":  prometheus.GetProjectName(),
			"swap_key": swapAddress,
		},
	}
	err := prometheus.ExamplePusherPush(log, labels)
	if err != nil {
		logger.Error("send msg to push_gateway failed!", logger.Errorv(err))
	}
}

// send 发送短信
func send(pool, currentAmountA, currentAmountB, needAmountA, needAmountB string) error {
	AccessKeyId := smsConfig.AccessKeyId
	AccessSecret := smsConfig.AccessSecret
	SignName := smsConfig.SignName
	Template := "SMS_245645033"
	smsClient, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", AccessKeyId, AccessSecret)

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = smsConfig.PhoneNumbers
	request.SignName = SignName
	request.TemplateCode = Template
	request.TemplateParam = fmt.Sprintf(`{"pool":"%s","currentAmountA":"%s","currentAmountB":"%s","needAmountA":"%s","needAmountB":"%s"}`,
		pool,
		currentAmountA,
		currentAmountB,
		needAmountA,
		needAmountB,
	)
	_, err = smsClient.SendSms(request)
	if err != nil {
		return err
	}
	return nil
}
