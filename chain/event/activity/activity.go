package event

import (
	"crypto/sha256"
	"encoding/base64"

	ag_solanago "github.com/gagliardetto/solana-go"
)

var (
	ActivityLayout map[string]Event = map[string]Event{
		"ClaimRewardEvent":      &ClaimRewardEvent{},
		"ClaimSecondPartyEvent": &ClaimSecondPartyEvent{},
	}

	ActivityDisc map[string]string = map[string]string{
		base64.StdEncoding.EncodeToString(EventDiscriminator("ClaimRewardEvent")):      "ClaimRewardEvent",
		base64.StdEncoding.EncodeToString(EventDiscriminator("ClaimSecondPartyEvent")): "ClaimSecondPartyEvent",
	}
)

func sighash(namespace string, name string) []byte {
	data := namespace + ":" + name
	sum := sha256.Sum256([]byte(data))
	return sum[0:8]
}

func EventDiscriminator(name string) []byte {
	return sighash("event", name)
}

type Event interface {
	GetUser() string
	GetMint() string
	GetAmount() uint64
	GetDegree() uint8
	GetCaffeine() uint64
	GetAmounts() []uint64
}

type ClaimRewardEvent struct {
	User     ag_solanago.PublicKey
	Mint     ag_solanago.PublicKey
	Amount   uint64
	Degree   uint8
	Caffeine uint64
}

func (o *ClaimRewardEvent) GetUser() string {
	return o.User.String()
}

func (o *ClaimRewardEvent) GetMint() string {
	return o.Mint.String()
}

func (o *ClaimRewardEvent) GetAmount() uint64 {
	return o.Amount
}

func (m *ClaimRewardEvent) GetDegree() uint8 {
	return m.Degree
}

func (m *ClaimRewardEvent) GetCaffeine() uint64 {
	return m.Caffeine
}

func (m *ClaimRewardEvent) GetAmounts() []uint64 {
	return []uint64{}
}

type ClaimSecondPartyEvent struct {
	User     ag_solanago.PublicKey
	Mint     ag_solanago.PublicKey
	Amounts  []uint64
	Degree   uint8
	Caffeine uint64
}

func (o *ClaimSecondPartyEvent) GetUser() string {
	return o.User.String()
}

func (o *ClaimSecondPartyEvent) GetMint() string {
	return o.Mint.String()
}

func (o *ClaimSecondPartyEvent) GetAmount() uint64 {
	return 0
}

func (m *ClaimSecondPartyEvent) GetDegree() uint8 {
	return m.Degree
}

func (m *ClaimSecondPartyEvent) GetCaffeine() uint64 {
	return m.Caffeine
}

func (m *ClaimSecondPartyEvent) GetAmounts() []uint64 {
	return m.Amounts
}
