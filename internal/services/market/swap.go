package handler

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"
	"git.cplus.link/go/akit/util/decimal"
	"git.cplus.link/go/akit/util/gquery"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/internal/worker/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

// SwapCount ...
func (t *MarketService) SwapCount(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)

	reply.SwapPairCount = market.GetSwapCountCache(args.TokenSwapAddress)

	if reply.SwapPairCount == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}

// GetTvl ...
func (t *MarketService) GetTvl(ctx context.Context, args *iface.GetTvlReq, reply *iface.GetTvlResp) error {
	defer rpcx.Recover(ctx)

	reply.Tvl = market.GetTvlCache()

	if reply.Tvl == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}

func (t *MarketService) GetTvlV2(ctx context.Context, args *iface.GetTvlReqV2, reply *iface.GetTvlRespV2) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	swapAddressList := make([]string, 0, 5)

	if args.SwapAddress == "" {
		// 表示获取所有swap address 的tvl
		for _, v := range sol.SwapConfigList() {
			swapAddressList = append(swapAddressList, v.SwapAccount)
		}
	} else {
		swapAddressList = append(swapAddressList, args.SwapAddress)
	}

	for _, swapAddress := range swapAddressList {
		tvl, err := t.redisClient.Get(ctx, domain.SwapTvlCountKey(swapAddress).Key).Result()
		if err != nil && !t.redisClient.ErrIsNil(err) {
			return errors.Wrap(err)
		} else if err != nil && !t.redisClient.ErrIsNil(err) {
			continue
		}

		tvlDecimal, _ := decimal.NewFromString(tvl)
		swapAddressTvl := &iface.SwapAddressTvl{
			SwapAccount: swapAddress,
			Tvl:         tvlDecimal,
		}

		reply.List = append(reply.List, swapAddressTvl)
	}

	// 直接查询数据库
	//swapTvl, err := model.GetLastSwapTvlCount(ctx, gquery.ParseQuery(args))
	//if err != nil {
	//	return errors.Wrap(err)
	//}

	//reply.SwapTvlCount = swapTvl

	// gateway 获取数据

	return nil
}

// GetNetRecord 某一时刻使用的rpc 网络情况，仅给后端使用
func (t *MarketService) GetNetRecord(ctx context.Context, args *iface.GetNetRecordReq, reply *iface.GetNetRecordResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	list, total, err := model.QueryNetRecords(ctx, limit(args.Limit), args.Offset, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total

	return nil
}

// Get24hVolV2 获取swap account 或者 user account 的24小时的交易量
func (t *MarketService) Get24hVolV2(ctx context.Context, args *iface.Get24hVolV2Req, reply *iface.Get24hVolV2Resp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	vol, err := t.redisClient.Get(ctx, domain.SwapVolCountLast24HKey(args.SwapAddress).Key).Result()
	if err != nil && !t.redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	} else if err == nil {

		aa := &model.SwapVol{}
		_ = json.Unmarshal([]byte(vol), aa)
		reply.Vol = aa.Vol

		return nil
	}

	// 在数据库里面找，并且同步到redis里 TODO

	tvlDecimal, _ := decimal.NewFromString(vol)

	reply.Vol = tvlDecimal

	return nil
}

func (t *MarketService) GetVolV2(ctx context.Context, args *iface.GetVolV2Req, reply *iface.GetVolV2Resp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	vol, err := t.redisClient.Get(ctx, domain.AccountSwapVolCountKey(args.SwapAddress).Key).Result()
	if err != nil && !t.redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	} else if err == nil {
		tvlDecimal, _ := decimal.NewFromString(vol)

		reply.Vol = tvlDecimal

		return nil
	}

	// 在数据库里面找，并且同步到redis里 TODO

	tvlDecimal, _ := decimal.NewFromString(vol)

	reply.Vol = tvlDecimal

	return nil
}

func (t *MarketService) QueryUserSwapTvlCount(ctx context.Context, args *iface.QueryUserSwapTvlCountReq, reply *iface.QueryUserSwapTvlCountResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	list, total, err := model.QueryUserSwapCount(ctx, limit(args.Limit), args.Offset, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total
	return nil
}

func (t *MarketService) QueryUserSwapTvlCountDay(ctx context.Context, args *iface.QueryUserSwapTvlCountDayReq, reply *iface.QueryUserSwapTvlCountDayResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	list, total, err := model.QueryUserSwapCountDay(ctx, limit(args.Limit), args.Offset, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total
	return nil
}
