package event

import (
	"encoding/base64"
	"strings"

	"git.cplus.link/go/akit/errors"
	"github.com/near/borsh-go"
)

var (
	LOG_START_INDEX int = len("Program log: ")
)

func NewActivityEventParser() EventParser {
	return EventParser{Discriminators: ActivityDisc, Layout: ActivityLayout}
}

type EventRep struct {
	User     string
	Mint     string
	Amount   uint64
	Degree   uint8
	Caffeine uint64
	Amounts  []uint64
	EventName string
}

type EventParser struct {
	Discriminators map[string]string
	Layout         map[string]Event
}

func (parser *EventParser) Decode(logMessages []string) ([]EventRep, error) {
	var events []EventRep
	for _, log := range logMessages {
		if !strings.HasPrefix(log, "Program log:") {
			continue
		}
		logArr, err := base64.StdEncoding.DecodeString(log[LOG_START_INDEX:])
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
				if err != nil {
					return []EventRep{}, errors.Wrap(err)
				}
				events = append(events, EventRep{
					User:     event.GetUser(),
					Mint:     event.GetMint(),
					Amount:   event.GetAmount(),
					Degree:   event.GetDegree(),
					Caffeine: event.GetCaffeine(),
					Amounts:  event.GetAmounts(),
					EventName: eventName,
				})
			}
		}
	}
	if len(events) == 0 {
		return []EventRep{}, nil
	}
	return events, nil
}
