package process

import (
	"context"
	"sync"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type ParserTransaction interface {
	GetSyncPoint() error
	WriteToDB(*domain.SwapTransaction) error
	ParserDate() error
}

// SyncKline sync transaction
type SyncKline struct {
	name       string
	spec       string
	swapConfig *domain.SwapConfig
}

func (s *SyncKline) Name() string {
	return s.name
}

func (s *SyncKline) GetSpecFunc() string {
	return s.spec
}

func (s *SyncKline) DeleteJobFunc(_ *JobInfo) error {
	return nil
}

func (s *SyncKline) Run() error {
	var err error
	swapPairBase, err := model.QuerySwapPairBase(context.TODO(), model.SwapAddress(s.swapConfig.SwapAccount))
	if err != nil {
		logger.Error("query swap_pair_bases err", logger.Errorv(err))
		return errors.Wrap(err)
	}
	if swapPairBase == nil {
		return nil
	}

	if swapPairBase.IsSync == false {
		return nil
	}

	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}

	swapAndUserCount := &SwapAndUserCount{
		LastTransactionID: lastSwapTransactionID,
		SwapAccount:       s.swapConfig.SwapAccount,
	}

	if err = swapAndUserCount.ParserDate(); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func CreateSyncKLine() error {
	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncKline", &m, func(value interface{}) JobInterface {
		return &SyncKline{
			name:       "sync_kline",
			spec:       getSpec("sync_kline"),
			swapConfig: value.(*domain.SwapConfig),
		}
	})
	if err != nil {
		return err
	}

	return nil
}
