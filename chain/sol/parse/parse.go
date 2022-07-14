package parse

import (
	"math"
	"strconv"

	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

type Tx struct {
	Data             rpc.GetTransactionResult
	SwapRecords      []*SwapRecord
	LiquidityRecords []*LiquidityRecord
	ClaimRecords     []*ClaimRecord
}

const cremaSwapProgramAddress = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"

var (
	swapConfigMap map[string]*domain.SwapConfig
)

func SetSwapConfig(configMap map[string]*domain.SwapConfig) {
	swapConfigMap = configMap
}

func NewTx(txData *domain.TxData) *Tx {
	return &Tx{
		Data:             rpc.GetTransactionResult(*txData),
		SwapRecords:      make([]*SwapRecord, 0),
		LiquidityRecords: make([]*LiquidityRecord, 0),
	}
}

// PrecisionConversion 精度转换
func PrecisionConversion(num decimal.Decimal, precision int) decimal.Decimal {
	return num.Div(decimal.NewFromFloat(math.Pow10(precision)))
}

func FormatFloat(num decimal.Decimal, dc int) decimal.Decimal {
	numf, _ := num.Float64()
	d := float64(1)
	if dc > 0 {
		d = math.Pow10(dc)
	}
	res := strconv.FormatFloat(math.Trunc(numf*d)/d, 'f', -1, 64)
	resf, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return num
	}

	return decimal.NewFromFloat(resf)
}

func FormatFloatCarry(num decimal.Decimal, dc int) decimal.Decimal {
	return num.Round(int32(dc))
}
