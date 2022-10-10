package event

import (
	"encoding/base64"
	"strings"

	"git.cplus.link/go/akit/errors"
	"github.com/near/borsh-go"
)

var (
	EventLayout = map[string]Event{
		SwapEventName:                            &SwapEvent{},
		IncreaseLiquidityEventName:               &IncreaseLiquidityEvent{},
		IncreaseLiquidityWithFixedTokenEventName: &IncreaseLiquidityWithFixedTokenEvent{},
		DecreaseLiquidityEventName:               &DecreaseLiquidityEvent{},
		CollectEventName:                         &CollectEvent{},
	}

	EventDisc = map[string]string{
		base64.StdEncoding.EncodeToString(EventDiscriminator(SwapEventName)):                            SwapEventName,
		base64.StdEncoding.EncodeToString(EventDiscriminator(IncreaseLiquidityEventName)):               IncreaseLiquidityEventName,
		base64.StdEncoding.EncodeToString(EventDiscriminator(IncreaseLiquidityWithFixedTokenEventName)): IncreaseLiquidityWithFixedTokenEventName,
		base64.StdEncoding.EncodeToString(EventDiscriminator(DecreaseLiquidityEventName)):               DecreaseLiquidityEventName,
		base64.StdEncoding.EncodeToString(EventDiscriminator(CollectEventName)):                         CollectEventName,
	}
)

type Event interface {
}

type EventParser struct {
	Discriminators map[string]string
	Layout         map[string]Event
}

type EventRep struct {
	EventName string
	Event     Event
}

type EventDecoder struct {
	Parsers *[]EventParser
}

var (
	LogStartStr       = "Program data: "
	LogStartIndex int = len(LogStartStr)
	eventDecoder  *EventDecoder
)

func NewEventParser() EventParser {
	return EventParser{Discriminators: EventDisc, Layout: EventLayout}
}

func NewEventDecoder() *EventDecoder {
	return &EventDecoder{Parsers: &[]EventParser{NewEventParser()}}
}

func (decoder *EventDecoder) Decode(logMessages []string) ([]EventRep, error) {
	var res []EventRep
	for _, parser := range *decoder.Parsers {
		event, err := parser.Decode(logMessages)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		res = append(res, event...)
	}
	return res, nil
}

func (parser *EventParser) Decode(logMessages []string) ([]EventRep, error) {
	var events []EventRep
	for _, log := range logMessages {
		if !strings.HasPrefix(log, LogStartStr) {
			continue
		}

		logArr, err := base64.StdEncoding.DecodeString(log[LogStartIndex:])
		if err != nil {
			continue
		}

		if len(logArr) < 9 {
			continue
		}

		eventPrefix := base64.StdEncoding.EncodeToString(logArr[:8])
		if eventName, exist := parser.Discriminators[eventPrefix]; !exist {
			continue
		} else {
			event := parser.Layout[eventName]
			if err := borsh.Deserialize(event, logArr[8:]); err == nil {
				events = append(events, EventRep{
					EventName: eventName,
					Event:     event,
				})
			}
		}
	}
	if len(events) == 0 {
		return []EventRep{}, nil
	}
	return events, nil
}

func Init() {
	eventDecoder = NewEventDecoder()
}

func GetEventDecoder() *EventDecoder {
	return eventDecoder
}
