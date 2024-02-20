package sol

import (
	"git.cplus.link/go/akit/errors"
	bin "github.com/gagliardetto/binary"
	"github.com/xmcontinue/solana-go"
)

type TickAccount struct {
	SwapVersion  uint8
	TokenSwapKey solana.PublicKey
	AccountType  uint8
	Len          int32
	// DataFlat     [][]byte
}

type Tick struct {
	Tick              int32
	TickPrice         decimalU12812
	LiquityGross      decimalU128
	LiquityNet        decimalU128
	FeeGrowthOutside0 decimalU12816
	FeeGrowthOutside1 decimalU12816
}

func GetTicks(data []byte) (*TickAccount, []Tick, error) {
	tickAccount := TickAccount{}

	err := bin.NewBinDecoder(data).Decode(&tickAccount)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	start := int32(PositionsHeadLen)

	ticks := []Tick{}
	for i := int32(0); i < tickAccount.Len; i++ {
		tick, end := Tick{}, start+84
		err = bin.NewBinDecoder(data[start:end]).Decode(&tick)
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
		ticks = append(ticks, tick)
		// tickAccount.DataFlat = append(tickAccount.DataFlat, data[start:end])

		start += 84
	}

	// for _, v := range positionsAccount.Positions {

	// i := big.Int{}
	// fmt.Println(v.Liquity.Val(), i.SetBytes(v.Liquity[:]).String())
	// fmt.Println(v.NftTokenId.String(), v.Liquity.Val(), v.FeeGrowthInsideALast.Val(), v.FeeGrowthInsideBLast.Val(), v.TokenAFee.Val(), v.TokenBFee.Val())
	// }

	return &tickAccount, ticks, nil
}
