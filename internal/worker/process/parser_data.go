package process

import (
	"context"
	"sync"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type parserIface interface {
	GetBeginId() (int64, error)
	GetParsingCutoffID() int64
	GetSwapAccount() string
	GetTransactions(limit, offset int, filters ...model.Filter) error
	ParserSwapInstruction() error
	ParserAllInstructionType() error
	UpdateLastTransActionID() error
}

type ParserTransaction interface {
	GetSyncPoint() error
	WriteToDB(*domain.SwapTransaction) error
	ParserDate() error
}

type WriteTyp struct {
	ID               int64
	SwapAccount      string
	BlockDate        *time.Time
	swapRecords      []parse.SwapRecordIface
	liquidityRecords []parse.LiquidityRecordIface
	claimRecords     []parse.CollectRecordIface
}

// swapKline sync transaction
type swapKline struct {
	name       string
	spec       string
	swapConfig *domain.SwapConfig
}

func (s *swapKline) Name() string {
	return s.name
}

func (s *swapKline) GetSpecFunc() string {
	return s.spec
}

func (s *swapKline) DeleteJobFunc(_ *JobInfo) error {
	return nil
}

func (s *swapKline) Run() error {
	var err error
	swapPairBase, err := model.QuerySwapPairBase(context.TODO(), model.SwapAddressFilter(s.swapConfig.SwapAccount))
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

	if s.swapConfig.Version == "v2" {
		swapV2 := &parserV2{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       s.swapConfig.SwapAccount,
			Version:           s.swapConfig.Version,
		}
		err = ParserSwapInstruction(swapV2)
	} else {
		swapCount := &parserV1{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       s.swapConfig.SwapAccount,
		}
		err = ParserSwapInstruction(swapCount)
	}

	if err != nil {
		logger.Error("parser data err", logger.Errorv(err))
	}

	return nil
}

func ParserSwapInstruction(iface parserIface) error {
	beginID, err := getBeginID(iface.GetSwapAccount())
	if err != nil {
		return errors.Wrap(err)
	}

	for {
		filters := []model.Filter{
			model.NewFilter("id <= ?", iface.GetParsingCutoffID()),
			model.SwapAddressFilter(iface.GetSwapAccount()),
			model.OrderFilter("id asc"),
			model.NewFilter("id > ?", beginID),
		}

		if err = iface.GetTransactions(100, 0, filters...); err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				break
			}
			return errors.Wrap(err)
		}

		if err = iface.ParserSwapInstruction(); err != nil {
			return errors.Wrap(err)
		}

		if err = iface.UpdateLastTransActionID(); err != nil {
			return errors.Wrap(err)
		}

		beginID, err = iface.GetBeginId()
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

// UserKline sync transaction
type UserKline struct {
	name       string
	spec       string
	swapConfig *domain.SwapConfig
}

func (s *UserKline) Name() string {
	return s.name
}

func (s *UserKline) GetSpecFunc() string {
	return s.spec
}

func (s *UserKline) DeleteJobFunc(_ *JobInfo) error {
	return nil
}

func (s *UserKline) Run() error {
	var err error
	swapPairBase, err := model.QuerySwapPairBase(context.TODO(), model.SwapAddressFilter(s.swapConfig.SwapAccount))
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

	if s.swapConfig.Version == "v2" {
		swapV2 := &parserV2{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       s.swapConfig.SwapAccount,
			Version:           s.swapConfig.Version,
		}
		err = ParserAllInstructionType(swapV2)
	} else {
		swapV1 := &parserV1{
			LastTransactionID: lastSwapTransactionID,
			SwapAccount:       s.swapConfig.SwapAccount,
		}
		err = ParserAllInstructionType(swapV1)
	}
	if err != nil {
		logger.Error("parser data err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil
}

func getUserMaxID(swapAccount string) (int64, error) {
	maxID, err := model.GetMaxUserCountKLineID(context.TODO(), swapAccount)
	if err != nil {
		if !errors.Is(err, errors.RecordNotFound) {
			return 0, errors.Wrap(err)
		}
		return 0, nil
	}
	return maxID, nil
}

func ParserAllInstructionType(iface parserIface) error {
	beginID, err := getUserMaxID(iface.GetSwapAccount())
	if err != nil {
		return errors.Wrap(err)
	}

	for {
		filters := []model.Filter{
			model.NewFilter("id <= ?", iface.GetParsingCutoffID()),
			model.SwapAddressFilter(iface.GetSwapAccount()),
			model.OrderFilter("id asc"),
			model.NewFilter("id > ?", beginID),
		}

		if err = iface.GetTransactions(100, 0, filters...); err != nil {
			if errors.Is(err, errors.RecordNotFound) {
				break
			}
			return errors.Wrap(err)
		}

		if err = iface.ParserAllInstructionType(); err != nil {
			return errors.Wrap(err)
		}

		if err = iface.UpdateLastTransActionID(); err != nil {
			return errors.Wrap(err)
		}

		beginID, err = iface.GetBeginId()
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

func CreateSyncKLine() error {
	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncSwapCountKline", &m, func(value interface{}) JobInterface {
		return &swapKline{
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

func CreateUserSyncKLine() error {
	m := sync.Map{}

	keys := sol.SwapConfigList()
	for _, v := range keys {
		m.Store(v.SwapAccount, v)
	}

	err := job.WatchJobForMap("SyncUserCountKLine", &m, func(value interface{}) JobInterface {
		return &UserKline{
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
