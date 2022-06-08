package handler

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"
	"git.cplus.link/go/akit/util/decimal"
	"git.cplus.link/go/akit/util/gquery"

	"git.cplus.link/crema/backend/chain/sol"
	solana "git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/internal/worker/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

// SwapCountOld ...
func (t *MarketService) SwapCountOld(ctx context.Context, args *iface.SwapCountReq, reply *iface.SwapCountOldResp) error {
	defer rpcx.Recover(ctx)

	reply.SwapPairCount = market.GetSwapCountCache(args.TokenSwapAddress)

	if reply.SwapPairCount == nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	return nil
}

// SwapCountList ...
func (t *MarketService) SwapCountList(ctx context.Context, args *iface.SwapCountListReq, reply *iface.SwapCountListResp) error {
	defer rpcx.Recover(ctx)

	list, err := model.QuerySwapCountKLines(
		ctx,
		100000,
		0,
		model.NewFilter("date_type = ?", args.DateType),
		model.NewFilter("date >= ?", args.BeginAt),
		model.NewFilter("date <= ?", args.EndAt),
	)

	if err != nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	for _, v := range list {
		reply.List = append(reply.List, &domain.SwapCountListInfo{
			TokenAUSDForContract: v.TokenAUSDForContract,
			TokenBUSDForContract: v.TokenBUSDForContract,
			TokenAVolume:         v.TokenAVolume,
			TokenBVolume:         v.TokenBVolume,
			TxNum:                uint64(v.TxNum),
			Date:                 v.Date,
			SwapAddress:          v.SwapAddress,
		})
	}
	return nil
}

// SwapCount ...
func (t *MarketService) SwapCount(ctx context.Context, _ *iface.NilReq, reply *iface.SwapCountResp) error {
	defer rpcx.Recover(ctx)

	res, err := t.redisClient.Get(ctx, domain.SwapTotalCountKey().Key).Result()
	if err != nil {
		return errors.Wrap(errors.RecordNotFound)
	}

	err = json.Unmarshal([]byte(res), reply)
	if err != nil {
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

	swapTvl, err := model.GetLastMaxTvls(ctx, gquery.ParseQuery(args))
	if err != nil {
		return errors.Wrap(err)
	}

	for _, tvl := range swapTvl {

		reply.List = append(reply.List, &iface.SwapAddressTvl{
			SwapAccount: tvl.SwapAddress,
			Tvl:         tvl.TokenABalance.Add(tvl.TokenBBalance),
		})
	}

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

	vol, err := t.redisClient.Get(ctx, domain.SwapVolCountLast24HKey(args.AccountAddress).Key).Result()
	if err != nil && !t.redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	} else if err == nil {

		aa := &model.SwapVol{}
		_ = json.Unmarshal([]byte(vol), aa)
		reply.SwapVol = aa

		return nil
	}

	// 未在redis中找到，直接返回空数据

	return nil
}

// GetVolV2 获取swap account 的总交易额
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

	// 在数据库里面找，直接返回空数据

	return nil
}

// func (t *MarketService) QueryUserSwapCount(ctx context.Context, args *iface.QueryUserSwapTvlCountReq, reply *iface.QueryUserSwapTvlCountResp) error {
//	defer rpcx.Recover(ctx)
//	if err := validate(args); err != nil {
//		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
//	}
//
//	list, total, err := model.QueryUserSwapCount(ctx, limit(args.Limit), args.Offset, gquery.ParseQuery(args))
//	if err != nil {
//		return errors.Wrap(err)
//	}
//
//	reply.Limit = limit(args.Limit)
//	reply.Offset = args.Offset
//	reply.List = list
//	reply.Total = total
//	return nil
// }

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

/*

Provide api for Activity project of nft metadata json, util the activity ends.
*/

func (t *MarketService) GetActivityNftMetadata(ctx context.Context, args *iface.GetActivityNftMetadataReq, reply *iface.GetActivityNftMetadataResp) error {
	defer rpcx.Recover(ctx)
	metadata, err := solana.GetMetadata(args.Mint)
	if err != nil {
		return errors.Wrap(err)
	}
	activityMeta, err := solana.GetActivityMeta(args.Mint)
	if err != nil {
		return errors.Wrap(err)
	}
	image := GetImageByDegree(activityMeta.Degree)
	reply.Name = metadata.Data.Name
	reply.Symbol = metadata.Data.Symbol
	reply.Image = image
	reply.SellerFeeBasisPoints = 0
	creators := []iface.Creator{}
	for _, e := range *metadata.Data.Creators {
		creators = append(creators, iface.Creator{
			Address: e.Address,
			Share:   e.Share,
		})
	}
	properties := &iface.Properties{
		Category: "image",
		Creators: &creators,
		Files: &[]iface.File{
			{
				Type: "image/png",
				Uri:  image,
			},
		},
	}
	reply.Properties = properties
	attributes := &[]iface.Attribute{
		{
			TraitType: "Level",
			Value:     GetLevelByDegree(activityMeta.Degree),
		},
	}
	reply.Attributes = attributes
	return nil
}

func GetImageByDegree(degree uint8) string {
	if degree == 1 {
		return "https://bafybeidt2nojsctionflvt7xpitxrkmdca6pycp3aft2uhld57uqafn2bq.ipfs.dweb.link/bronze.png"
	}
	if degree == 2 {
		return "https://bafybeihpr3kgcpksief7o2snebbfu65vc3jykimcyjjvyonyauzxbfunpu.ipfs.dweb.link/silver.png"
	}
	if degree == 3 {

		return "https://bafybeiateouoesl4s2f7ju2qbggcrifsst3iajjh6zjz5gpogx4j6andj4.ipfs.dweb.link/Gold.png"
	}
	if degree == 4 {
		return "https://bafybeicnen5re24nye47fplvrcivbwms6fv5v5qp4oqmhqwligvgimr5he.ipfs.dweb.link/platinum.png"
	}
	if degree == 5 {
		return "https://bafybeiedxet5oez2j6epby2nxkyzgsbllflwwwrpzny37riie2iljzgydi.ipfs.dweb.link/diamond.png"
	}
	return ""
}

func GetLevelByDegree(degree uint8) string {
	if degree == 1 {
		return "Bronze"
	}
	if degree == 2 {
		return "Silver"
	}
	if degree == 3 {
		return "Gold"
	}
	if degree == 4 {
		return "Platinum"
	}
	if degree == 5 {
		return "Diamond"

	}
	return ""
}
