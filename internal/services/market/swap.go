package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"
	"git.cplus.link/go/akit/util/gquery"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/internal/worker/market"

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

	swapTvl, err := model.GetLastSwapTvlCount(ctx, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	reply.SwapTvlCount = swapTvl

	// gateway 获取数据

	return nil
}

// GetNetRecord 某一时刻使用的rpc 网络情况
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

func (t *MarketService) Get24hVolV2(ctx context.Context, args *iface.Get24hVolV2Req, reply *iface.Get24hVolV2Resp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	list, total, err := model.QuerySwapTvlCountDay(ctx, limit(args.Limit), args.Offset, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total

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
