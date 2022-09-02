package parse

import (
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	event "git.cplus.link/crema/backend/chain/event/parser"
	"git.cplus.link/crema/backend/pkg/domain"
)

const cremaSwapProgramAddressV2 = "CcLs6shXAUPEi19SGyCeEHU9QhYAWzV2dRpPPNA4aRb7"

type SwapRecordIface interface {
	GetSwapConfig() *domain.SwapConfig
	GetUserOwnerAccount() string
	GetPrice() decimal.Decimal
	GetTokenAVolume() decimal.Decimal
	GetTokenBVolume() decimal.Decimal
	GetTokenABalance() decimal.Decimal
	GetTokenBBalance() decimal.Decimal
	GetDirection() int8
}

type Txv2 struct {
	SwapRecords      []*SwapRecordV2
	LiquidityRecords []*LiquidityRecordV2
	ClaimRecords     []*CollectRecordV2
}

func NewTxV2() *Txv2 {
	return &Txv2{
		SwapRecords:      make([]*SwapRecordV2, 0),
		LiquidityRecords: make([]*LiquidityRecordV2, 0),
	}
}

func (t *Txv2) ParseAllV2(logs string) error {
	var logList []string
	err := json.Unmarshal([]byte(logs), &logList)
	if err != nil {
		return errors.Wrap(err)
	}
	logMessageEvents, err := event.GetEventDecoder().Decode(logList)
	if err != nil {
		return errors.Wrap(err)
	}

	//没有解析到event 过滤
	if len(logMessageEvents) == 0 {
		return nil
	}

	// 解析事件，
	for _, logMessageEvent := range logMessageEvents {
		if logMessageEvent.EventName == event.SwapEventName {
			err = t.createSwapRecord(logMessageEvent)
		} else if logMessageEvent.EventName == event.IncreaseLiquidityEventName {
			err = t.createIncreaseLiquidityRecord(logMessageEvent)
		} else if logMessageEvent.EventName == event.DecreaseLiquidityEventName {
			err = t.createDecreaseLiquidityRecord(logMessageEvent)
		} else if logMessageEvent.EventName == event.CollectEventName {
			err = t.createCollectRecord(logMessageEvent)
		} else {
			continue
		}

		if err != nil {
			continue
		}
	}

	return nil
}

func (t *Txv2) ParseSwapV2(logs string) error {
	var logList []string
	err := json.Unmarshal([]byte(logs), &logList)
	if err != nil {
		return errors.Wrap(err)
	}
	logMessageEvents, err := event.GetEventDecoder().Decode(logList)
	if err != nil {
		return errors.Wrap(err)
	}

	//没有解析到event 过滤
	if len(logMessageEvents) == 0 {
		return nil
	}

	// 解析事件，
	for _, logMessageEvent := range logMessageEvents {
		// 直接洗swap事件
		if logMessageEvent.EventName == event.SwapEventName {
			err = t.createSwapRecord(logMessageEvent)
			if err != nil {
				continue
			}
		}
	}

	return nil
}
