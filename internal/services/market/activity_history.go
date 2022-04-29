package handler

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/iface"
)

var Decimals = map[string]uint64{"Crm": 1000000, "Marinade": 1000000, "Port": 1000000, "Hubble": 1000000000, "Nirv": 1000000}

func (t *MarketService) CreateActivityHistory(ctx context.Context, args *iface.CreateActivityHistoryReq, reply *iface.CreateActivityHistoryResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}
	return nil
}

func (t *MarketService) GetActivityHistoryByUser(ctx context.Context, args *iface.GetActivityHistoryByUserReq, reply *iface.GetActivityHistoryByUserResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	history, err := model.SelectByUser(ctx, args.User)
	if err != nil {
		return errors.Wrap(err)
	}
	historys := []*iface.ActivityHistoryResp{}
	for _, e := range history {
		historys = append(historys, &iface.ActivityHistoryResp{
			ID:           e.ID,
			CreatedAt:    e.CreatedAt,
			UpdatedAt:    e.UpdatedAt,
			UserKey:      e.UserKey,
			MintKey:      e.MintKey,
			Crm:          float64(e.Crm / Decimals["Crm"]),
			Marinade:     float64(e.Marinade / Decimals["Marinade"]),
			Port:         float64(e.Port / Decimals["Port"]),
			Hubble:       float64(e.Hubble / Decimals["Hubble"]),
			Nirv:         float64(e.Nirv / Decimals["Nirv"]),
			SignatureCrm: e.SignatureCrm,
			Signature:    e.Signature,
			BlockTime:    e.BlockTime,
			Degree:       e.Degree,
			Caffeine:     e.Caffeine / 1000000,
		})
	}
	reply.History = historys
	return nil
}
