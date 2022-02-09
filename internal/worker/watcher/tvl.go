package watcher

import (
	"sync"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
)

// SyncTvl 同步Tvl
type SyncTvl struct {
	name string
	spec string
	tvl  *sol.TVL
}

func (s *SyncTvl) Name() string {
	return s.name
}

func (s *SyncTvl) GetSpecFunc() string {
	return s.spec
}

func (s *SyncTvl) DeleteJobFunc(_ *JobInfo) error {
	return nil
}

func (s *SyncTvl) Run() error {

	err := s.tvl.Start()

	if err != nil {

		return errors.Wrap(err)

	}

	return nil
}

func CreateSyncTvl() error {
	// 同步单个tvl

	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncTvl", &m, func(value interface{}) JobInterface {
		return &SyncTvl{
			name: "sync_tvl",
			spec: "0 */10 * * * *",
			tvl:  sol.NewTVL(value.(*sol.SwapConfig)),
		}
	})
	if err != nil {
		return err
	}

	return nil
}
