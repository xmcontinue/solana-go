package parse

import (
	"math"
	"math/big"
	"strconv"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	sDecimal "github.com/shopspring/decimal"

	"git.cplus.link/go/akit/util/decimal"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

var MaxInt64 *big.Float

func init() {
	MaxInt64 = big.NewFloat(1)
	for i := 0; i < 64; i++ {
		MaxInt64 = MaxInt64.Mul(MaxInt64, big.NewFloat(2))
	}
}

type Tx struct {
	Data             rpc.GetTransactionResult
	SwapRecords      []*SwapRecord
	LiquidityRecords []*LiquidityRecord
	ClaimRecords     []*ClaimRecord
	TransAction      *solana.Transaction
}

type TxIface interface {
	GetTx() *Tx
}

func (t *Tx) GetTx() *Tx {
	return t
}

const cremaSwapProgramAddress = "6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319"

var (
	swapConfigMap map[string]*domain.SwapConfig
)

func SetSwapConfig(configMap map[string]*domain.SwapConfig) {
	swapConfigMap = configMap
}

// 使用前提是 不会溢出
func int64ListTouInt16List(int16List []int64) []uint16 {
	int64List := make([]uint16, 0, len(int16List))
	for _, v := range int16List {
		int64List = append(int64List, uint16(v))
	}
	return int64List
}

// NewTx 需要兼容JSON和base64拉取的数据
func NewTx(txData *domain.TxData) *Tx {
	var tx *solana.Transaction
	var err error
	tx, err = solana.TransactionFromDecoder(bin.NewBinDecoder(rpc.GetTransactionResult(*txData).Transaction.GetBinary()))
	if err != nil {
		// 说明是通过json格式拉取的数据，在这里统一转换成solana.Transaction
		ParsedTransaction, err := rpc.GetTransactionResult(*txData).Transaction.GetTransaction()
		if err != nil {
			logger.Error("TransactionFromDecoder：mismatched data format", logger.Errorv(err))
			return nil
		}
		tx = &solana.Transaction{
			Signatures: ParsedTransaction.Signatures,
			Message: solana.Message{
				AccountKeys: ParsedTransaction.Message.AccountKeys,
				Instructions: func() []solana.CompiledInstruction {
					instructions := make([]solana.CompiledInstruction, 0, len(ParsedTransaction.Message.Instructions))
					for _, v := range ParsedTransaction.Message.Instructions {
						instructions = append(instructions, solana.CompiledInstruction{
							ProgramIDIndex: v.ProgramIDIndex,
							Accounts:       v.Accounts,
							Data:           v.Data,
						})
					}
					return instructions
				}(),
			},
		}
	}

	return &Tx{
		Data:             rpc.GetTransactionResult(*txData),
		SwapRecords:      make([]*SwapRecord, 0),
		LiquidityRecords: make([]*LiquidityRecord, 0),
		TransAction:      tx,
	}
}

// ParseTxALl 解析TX内的所有类型
func (t *Tx) ParseTxALl() error {
	if t.ParseTxToSwap() != nil && t.ParseTxToLiquidity() != nil && t.ParseTxToClaim() != nil {
		return errors.RecordNotFound
	}
	return nil
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

func BankToString(v decimal.Decimal, r int32) string {
	f, _ := v.Float64()
	return sDecimal.NewFromFloat(f).RoundFloor(r).String()
}

func Bank(v decimal.Decimal, r int32) decimal.Decimal {
	f, _ := v.Float64()
	s := sDecimal.NewFromFloat(f).RoundFloor(r).String()
	d, _ := decimal.NewFromString(s)
	return d
}
